package admissionreview

import (
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ValidateResult is the returned result from the validation process.
type ValidateResult struct {
	// Status gives detailed information in the case of failure.
	// +optional
	Status *metav1.Status
	// Allow determines whether to allow the given API request at all.
	Allow bool
}

func (result *ValidateResult) admissionResponse(uid types.UID) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		UID:     uid,
		Allowed: result.Allow,
		Result:  result.Status,
	}
}

// ResourceValidator receives the raw request JSON representation as []byte. Unmarshalls this and returns the extracted request object.
// Furthermore, relevant modifications are applied and the modified response object returned.
// Errors should be handled internally and modify the resulting ValidateResult accordingly.
type ResourceValidator[T any] func(request *T) *ValidateResult

// ValidatingReviewer is the implementation of the ReviewerHandler interface. Checks the GroupVersionKind of the receives request
// against what the given reviewer.Modifier supports. A miss match will result in a non-modifying response and
// the allow value set to the value given by reviewer.AllowOnModifierMiss.
// Otherwise the Patch function of the Modifier interface is called, a JSON Patch is constructed from the result
// and wrapped into an admissionResponse.
func ValidatingReviewer[T any](validator ResourceValidator[T], compatibleGroupVersionKinds ...*metav1.GroupVersionKind) ReviewerHandler {
	return ReviewFunc(func(arRequest *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
		request, skipValidate := UnmarshallAdmissionRequest[T](arRequest.Object.Raw, compatibleGroupVersionKinds, &arRequest.Kind)
		if skipValidate != nil {
			return skipValidate.admissionResponse(arRequest.UID)
		}
		return validator(request).admissionResponse(arRequest.UID)
	})
}
