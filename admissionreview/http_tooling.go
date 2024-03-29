// Package admissionreview provides methods to handle Kubernetes admission review requests for webhook microservices
package admissionreview

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type httpError struct {
	// underlying error
	Err error
	// suggested HTTP (error) status code
	HttpResponseStatus int
}

// Handle receives a Reviewer interface and the ResponseWriter and Request from the http.Handler interface.
// This covers the IO part as well as error logging, HTTP response code handling and the construction
// of the AdmissionReview response object.
// Do not use if you do not wish to use zerolog for logging. GetAdmissionReviewFromHttp is an alternative that
// provides the relevant IO handling toolings and let the caller handle the HTTP and logging part.
func Handle(reviewer Reviewer, w http.ResponseWriter, r *http.Request) {
	arReview, httpErr := getAdmissionReviewFromHttp(r)
	if httpErr != nil {
		log.Error().Err(httpErr.Err).Msg("Error during request parsing")
		w.WriteHeader(httpErr.HttpResponseStatus)
		return
	}
	response := reviewer.Review(arReview.Request)

	// actually call the admission reviewer and return the response
	arResponse := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: response,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(&arResponse)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode response")
		// try to adjust response, depends on the error details if this has an effect
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// getAdmissionReviewFromHttp receives a HTTP request and handles the IO and unmarshal part
// to extract the AdmissionReview object from it.
func getAdmissionReviewFromHttp(r *http.Request) (*admissionv1.AdmissionReview, *httpError) {
	if r.Method != http.MethodPost {
		return nil, &httpError{fmt.Errorf("unsupported HTTP method: %v", r.Method), http.StatusMethodNotAllowed}
	}
	if r.Body == nil {
		return nil, &httpError{errors.New("body missing"), http.StatusBadRequest}
	}
	var arReview admissionv1.AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&arReview); err != nil {
		return nil, &httpError{fmt.Errorf("failed to read and unmarshal body: %w", err), http.StatusBadRequest}
	}
	return &arReview, nil
}

// UnmarshallAdmissionRequest checks if the requestGroupVersionKind fits to the provided selector and unmarshalls the raw request into a the result pointer if this is the case.
// The presence of the validateResult implies that the skip condition has been fulfilled (Allow is true) or an error occurred during unmarshalling (Allow is false and Status contains the error).
func UnmarshallAdmissionRequest[T any](rawRequest []byte, compatibleGroupVersionKinds []*metav1.GroupVersionKind, requestGroupVersionKind *metav1.GroupVersionKind) (request *T, validateResult *ValidateResult) {
	if !Contains(compatibleGroupVersionKinds, requestGroupVersionKind) {
		return nil, &ValidateResult{
			Allow: true,
		}
	}
	var result T
	if err := json.Unmarshal(rawRequest, &result); err != nil {
		return nil, &ValidateResult{
			Allow:  false,
			Status: GetErrorStatus(http.StatusUnprocessableEntity, "failed to unmarshal into namespace object", err),
		}
	}
	return &result, nil
}
