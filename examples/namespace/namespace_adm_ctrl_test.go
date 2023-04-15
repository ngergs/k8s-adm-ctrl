package main

import (
	_ "embed"
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/ngergs/k8s-adm-ctrl/admissionreview"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

//go:embed testdata/request_invalid.json
var requestInvalid string

//go:embed testdata/request_valid.json
var requestValid string

// TestValidationSuccess tests that valid requests are allowed to pass
func TestValidationSuccess(t *testing.T) {
	resp := testValidation(t, requestValid)
	require.True(t, resp.Response.Allowed)
}

// TestValidationFailure tests that invalid requests are rejected with an UnprocessableEntity status code
func TestValidationFailure(t *testing.T) {
	resp := testValidation(t, requestInvalid)
	require.False(t, resp.Response.Allowed)
	require.Equal(t, "Failure", resp.Response.Result.Status)
	require.Equal(t, http.StatusUnprocessableEntity, int(resp.Response.Result.Code))
}

// TestMutationNoChange checks that if no mutation is necessary the request is allowed
func TestMutationNoChange(t *testing.T) {
	resp := testValidation(t, requestValid)
	require.True(t, resp.Response.Allowed)
	require.Nil(t, resp.Response.Patch)
}

// TestMutationChange checks that the mutating reviewer does not allow invalid requests.
// Furthermore, the returned JSON patch is applied onto the original invalid request and the validity of the result verified.
func TestMutationChange(t *testing.T) {
	req := requestInvalid
	resp := testMutation(t, req)
	require.True(t, resp.Response.Allowed)
	require.NotNil(t, resp.Response.Patch)
	patch, err := jsonpatch.DecodePatch(resp.Response.Patch)
	require.NoError(t, err)
	// verify that the original Request would be rejected
	resp = testValidation(t, requestInvalid)
	require.False(t, resp.Response.Allowed)
	// patch and try again
	req = applyJsonPatch(t, req, patch)
	resp = testValidation(t, req)
	require.True(t, resp.Response.Allowed)
}

// applyJsonPatch applies a RFC6902 JSON Patch onto the .Request.Object of the input JSON encoded AdmissionReview
func applyJsonPatch(t *testing.T, admReqStr string, patch jsonpatch.Patch) string {
	var admReq admissionv1.AdmissionReview
	err := json.Unmarshal([]byte(admReqStr), &admReq)
	require.NoError(t, err)
	objJson, err := json.Marshal(admReq.Request.Object)
	require.NoError(t, err)
	modifiedJson, err := patch.Apply(objJson)
	require.NoError(t, err)
	err = json.Unmarshal(modifiedJson, &admReq.Request.Object)
	require.NoError(t, err)
	modifiedReq, err := json.Marshal(admReq)
	require.NoError(t, err)
	return string(modifiedReq)
}

// testValidation creates a namespace validation reviewer, calls it with the serialized request and returns the response
func testValidation(t *testing.T, req string) *admissionv1.AdmissionReview {
	adm := namespaceLabelMutater{}
	rev := admissionreview.ValidatingReviewer(adm.Validate, compatibleGroupVersionKind)
	return testAdmissionReview(t, req, rev)
}

// testMutation creates a namespace mutation reviewer, calls it with the serialized request and returns the response
func testMutation(t *testing.T, req string) *admissionv1.AdmissionReview {
	adm := namespaceLabelMutater{}
	rev := admissionreview.MutatingReviewer(adm.Patch, compatibleGroupVersionKind)
	return testAdmissionReview(t, req, rev)
}

// testAdmissionReview takes a serialized admissionRequest, calls the reviewer http.Handler and returns the deserialized response
func testAdmissionReview(t *testing.T, req string, rev http.Handler) *admissionv1.AdmissionReview {
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(req))
	require.NoError(t, err)
	rev.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Result().StatusCode)
	defer w.Result().Body.Close()
	var resp admissionv1.AdmissionReview
	err = json.NewDecoder(w.Result().Body).Decode(&resp)
	require.NoError(t, err)
	return &resp
}
