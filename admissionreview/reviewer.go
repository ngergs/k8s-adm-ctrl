package admissionreview

import (
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reviewer receives a Kubernetes AdmissionRequest and returns the corresponding AdmissionResponse
// Errors should be handled internally and modify the resulting AdmissionResponse accordingly
type Reviewer interface {
	Review(*admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse
}

// ReviewerHandler combines the Reviewer and http.Handler interfaces. Used for functions which provides
// a reviewer combined with an already setup handler for easy use in combination with the http package.
type ReviewerHandler interface {
	Reviewer
	http.Handler
}

// Reviewer receives a Kubernetes AdmissionRequest and returns the corresponding AdmissionResponse
// Errors should be handled internally and modify the resulting AdmissionResponse accordingly.
// Implements the reviewer and http.Handler interface.
type reviewFuncWrapper struct {
	reviewFunc func(*admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse
}

func (reviewer *reviewFuncWrapper) Review(arRequest *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	return reviewer.reviewFunc(arRequest)
}

func (reviewer *reviewFuncWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Handle(reviewer, w, r)
}

// ReviewFunc is a helper function to wrap a review function into a corresponding object
func ReviewFunc(reviewFunc func(*admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse) ReviewerHandler {
	return &reviewFuncWrapper{reviewFunc: reviewFunc}
}

// GetErrorStatus receives a suggested HTTP (error) status code, an error description as well as
// an underlying error and constructs a Failure metav1.Status from this information
func GetErrorStatus(httpStatus int32, errDiscription string, err error) *metav1.Status {
	return &metav1.Status{
		Status:  "Failure",
		Message: fmt.Sprintf("%s: %v", errDiscription, err),
		Code:    httpStatus,
	}
}

// Contains checks if the obj argument is contained in the slice argument
func Contains(slice []metav1.GroupVersionKind, obj metav1.GroupVersionKind) bool {
	for _, el := range slice {
		if el == obj {
			return true
		}
	}
	return false
}
