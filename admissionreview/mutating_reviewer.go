package admissionreview

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/types"
	"net/http"

	"github.com/wI2L/jsondiff"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var jsonPatchType = admissionv1.PatchTypeJSONPatch

// Patch is used to construct the relevant JSON Patch operations.
type Patch[T any] struct {
	// Request is the unmarshalled original request object. Returning nil here will yield an empty JSON patch response.
	Request *T
	// Response is the modified request object. Returning nil here will yield an empty JSON patch response.
	Response *T
}

// ResourceMutater receives the raw request JSON representation as []byte. Unmarshalls this and returns the extracted request object.
// Furthermore, relevant modifications are applied and the modified response object returned.
// The patches struct pointer might be nil. If it is present all patches have to be processed for the validate result to hold.
type ResourceMutater[T any] func(request *T) (*ValidateResult, *Patch[T])

// MutatingReviewer is the implementation of the ReviewerHandler interface. Checks the GroupVersionKind of the receives request
// against what the given reviewer.Modifier supports. A miss match will result in a non-modifying response and
// the allow value set to the value given by reviewer.AllowOnModifierMiss.
// Otherwise the Patch function of the Modifier interface is called, a JSON Patch is constructed from the result
// and wrapped into an admissionResponse.
func MutatingReviewer[T any](mutater ResourceMutater[T], compatibleGroupVersionKinds ...*metav1.GroupVersionKind) ReviewerHandler {
	return ReviewFunc(func(arRequest *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
		var request T
		skipMutateResult := UnmarshallAdmissionRequest(&request, arRequest.Object.Raw, compatibleGroupVersionKinds, &arRequest.Kind)
		if skipMutateResult != nil {
			return skipMutateResult.admissionResponse(arRequest.UID)
		}
		result, patches := mutater(&request)
		if !result.Allow || patches == nil {
			return result.admissionResponse(arRequest.UID)
		}

		// collect changes into JSON Patch
		patch, err := jsondiff.Compare(patches.Request, patches.Response)
		if err != nil {
			return jsonPatchErrorResponse(arRequest.UID, err)
		}
		patchJson, err := json.Marshal(&patch)
		if err != nil {
			return jsonMarshallErrorResponse(arRequest.UID, err)
		}
		// everything has worked, construct response
		response := result.admissionResponse(arRequest.UID)
		response.Patch = patchJson
		response.PatchType = &jsonPatchType
		return response
	})
}

func jsonPatchErrorResponse(uid types.UID, err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		UID:     uid,
		Allowed: false,
		Result:  GetErrorStatus(http.StatusInternalServerError, "failed to create JSON patch request and supposed response object", err),
	}
}

func jsonMarshallErrorResponse(uid types.UID, err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		UID:     uid,
		Allowed: false,
		Result:  GetErrorStatus(http.StatusInternalServerError, "failed to marshall JSON patch", err),
	}
}
