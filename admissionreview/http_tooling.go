// provides methods to handle Kubernetes admission review requests for webhook microservices
package admissionreview

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

// ToHandlFunc receives a Reviewer interface and wraps this into a HTTP handler.
// This covers the IO part as well as error logging, HTTP response code handling and the construction
// of the AdmissionReview response object.
// Do not use if you do not wish to use zerolog forlogging. GetAdmissionReviewFromHttp is an alternative that
// provides the relevant IO handling toolings and let the caller handle the HTTP and logging part.
func ToHandelFunc(reviewer Reviewer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(&arResponse)
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
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, &httpError{fmt.Errorf("failed to read body: %w", err), http.StatusInternalServerError}
	}
	var arReview admissionv1.AdmissionReview
	if err := json.Unmarshal(data, &arReview); err != nil {
		return nil, &httpError{fmt.Errorf("failed to unmarshal body: %w", err), http.StatusBadRequest}
	}
	return &arReview, nil
}

// UnmarshallAdmissionRequestchecks if the requestGroupVersionKind fits to the provided compatibleGroupVersionKinds and unmarshalls the raw request into a the result pointer if this is the case.
// The presence of the validateResult implies that the skip condition has been fulfilled (Allow is true) or an error occured during unmarshalling (Allow is false and Status contains the error).
func UnmarshallAdmissionRequest(compatibleGroupVersionKinds []metav1.GroupVersionKind, result interface{},
	requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *ValidateResult {
	if !Contains(compatibleGroupVersionKinds, *requestGroupVersionKind) {
		return &ValidateResult{
			Allow: true,
		}
	}
	if err := json.Unmarshal(rawRequest, result); err != nil {
		return &ValidateResult{
			Allow:  false,
			Status: GetErrorStatus(http.StatusUnprocessableEntity, "failed to unmarshal into namespace object", err),
		}
	}
	return nil
}
