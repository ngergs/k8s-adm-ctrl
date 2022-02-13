package admissionreview

import (
	"encoding/json"
	"net/http"

	"github.com/wI2L/jsondiff"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Patch is used to construct the relevant JSON Patch operations.
type Patch struct {
	// Request is the unmarshalled original request object. Returning nil here will yield an empty JSON patch response.
	Request interface{}
	// Response is the modified request object. Returning nil here will yield an empty JSON patch response.
	Response interface{}
}

// ResourceMutater receives the raw request JSON representation as []byte. Unmarshalls this and returns the extracted request object.
// Furthermore, relevant modifications are applied and the modified response object returned.
// The patches struct pointer might be nil. If it is present all patches have to be processed for the validate result to hold.
type ResourceMutater func(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) (*ValidateResult, *Patch)

// Review is the implementation of the Reviewer interface. Checks the GroupVersionKind of the receives request
// against what the given reviewer.Modifier supports. A miss match will result in a non-modifying response and
// the allow value set to the value given by reviewer.AllowOnModifierMiss.
// Otherwise the Patch function of the Modifier interface is called, a JSON Patch is constructed from the result
// and wrapped into an AdmissionResponse.
func MutatingReviewer(mutater ResourceMutater) ReviewerHandler {
	return ReviewFunc(func(arRequest *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
		result, patches := mutater(&arRequest.Kind, arRequest.Object.Raw[:])
		if !result.Allow || patches == nil {
			return &admissionv1.AdmissionResponse{
				UID:     arRequest.UID,
				Allowed: result.Allow,
				Result:  result.Status,
			}
		}

		// collect changes into JSON Patch
		patch, err := jsondiff.Compare(patches.Request, patches.Response)
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
			Allowed:   result.Allow,
			PatchType: &patchType,
			Patch:     patchJson,
			Result:    result.Status,
		}
	})
}
