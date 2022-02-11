package admissionreview

import (
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source=$GOFILE -destination=../${GOPACKAGE}_tests/mock_${GOFILE} -package=${GOPACKAGE}_tests
type ValidateResult struct {
	// Allow determines whether to allow the given API request at all.
	Allow bool
	// Status gives detailed information in the case of failure.
	// +optional
	Status *metav1.Status
}

// ResourceValidator receives the raw request JSON representation as []byte. Unmarshalls this and returns the extracted request object.
// Furthermore, relevant modifications are applied and the modified response object returned.
// Errors should be handled internally and modify the resulting ValidateResult accordingly.
type ResourceValidator func(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *ValidateResult

// Review is the implementation of the Reviewer interface. Checks the GroupVersionKind of the receives request
// against what the given reviewer.Modifier supports. A miss match will result in a non-modifying response and
// the allow value set to the value given by reviewer.AllowOnModifierMiss.
// Otherwise the Patch function of the Modifier interface is called, a JSON Patch is constructed from the result
// and wrapped into an AdmissionResponse.
func ValidatingReviewer(validator ResourceValidator) Reviewer {
	return ReviewFunc(func(arRequest *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
		result := validator(&arRequest.Kind, arRequest.Object.Raw[:])

		return &admissionv1.AdmissionResponse{
			UID:     arRequest.UID,
			Allowed: result.Allow,
			Result:  result.Status,
		}
	})
}
