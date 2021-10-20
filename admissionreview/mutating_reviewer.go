package admissionreview

import (
	"encoding/json"
	"net/http"

	"github.com/wI2L/jsondiff"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PatchResult struct {
	// Allow determines whether to allow the given API request at all.
	// Settings this to false means Request and Response should be ignored.
	Allow bool
	// Request is the unmarshalled original request object. Returning nil here will yield an empty JSON patch response.
	// +optional
	Request interface{}
	// Response is the modified request object. Returning nil here will yield an empty JSON patch response.
	// +optional
	Response interface{}
	// Status gives detailed information in the case of failure.
	// +optional
	Status *metav1.Status
}

//go:generate mockgen -source=$GOFILE -destination=../${GOPACKAGE}_tests/mock_${GOFILE} -package=${GOPACKAGE}_tests
type ResourceMutater interface {
	// Patch receives the raw request JSON representation as []byte. Unmarshalls this and returns the extracted request object.
	// Furthermore, relevant modifications are applied and the modified response object returned.
	// Errors should be handled internally and modify the resulting AdmissionResponse accordingly
	Patch(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *PatchResult
}

type MutatingReviewer struct {
	// Mutater holds the actual object modification logic
	Mutater ResourceMutater
}

// Review is the implementation of the Reviewer interface. Checks the GroupVersionKind of the receives request
// against what the given reviewer.Modifier supports. A miss match will result in a non-modifying response and
// the allow value set to the value given by reviewer.AllowOnModifierMiss.
// Otherwise the Patch function of the Modifier interface is called, a JSON Patch is constructed from the result
// and wrapped into an AdmissionResponse.
func (reviewer *MutatingReviewer) Review(arRequest *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	patchResult := reviewer.Mutater.Patch(&arRequest.Kind, arRequest.Object.Raw[:])
	if !patchResult.Allow || patchResult.Request == nil || patchResult.Response == nil {
		return &admissionv1.AdmissionResponse{
			UID:     arRequest.UID,
			Allowed: patchResult.Allow,
			Result:  patchResult.Status,
		}
	}

	// collect changes into JSON Patch if both are given
	patch, err := jsondiff.Compare(patchResult.Request, patchResult.Response)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			UID:     arRequest.UID,
			Allowed: false,
			Result:  GetErrorStatus(http.StatusInternalServerError, "failed to create JSON patch request and supposed response object", err),
		}
	}
	patchJson, err := json.Marshal(&patch)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			UID:     arRequest.UID,
			Allowed: false,
			Result:  GetErrorStatus(http.StatusInternalServerError, "failed to marshall JSON patch", err),
		}
	}
	// construct response
	patchType := admissionv1.PatchType(admissionv1.PatchTypeJSONPatch)
	return &admissionv1.AdmissionResponse{
		UID:       arRequest.UID,
		Allowed:   patchResult.Allow,
		PatchType: &patchType,
		Patch:     patchJson,
		Result:    patchResult.Status,
	}
}
