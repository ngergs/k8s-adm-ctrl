package admissionreview

import (
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source=$GOFILE -destination=../${GOPACKAGE}_tests/mock_${GOFILE} -package=${GOPACKAGE}_tests
type Reviewer interface {
	// Review receives a Kubernetes AdmissionRequest and returns the corresponding AdmissionResponse
	// Errors should be handled internally and modify the resulting AdmissionResponse accordingly
	Review(*admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse
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
