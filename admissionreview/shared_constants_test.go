package admissionreview_test

import (
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type dataType struct {
	Test  string `json:"test,omitempty"`
	Test2 string `json:"test2,omitempty"`
}

var data = []byte("{\"test\":\"123\"}")
var dataMutated = []byte("{\"test\":\"123\", \"test2\":\"234\"}")
var dataPatch = []byte("[{\"value\":\"234\",\"op\":\"add\",\"path\":\"/test2\"}]")

var groupVersionKind = &metav1.GroupVersionKind{
	Group:   "",
	Version: "v1",
	Kind:    "Namespace",
}
var status = &metav1.Status{
	Status:  "failure",
	Message: "test",
}
var arRequest = &admissionv1.AdmissionRequest{
	UID:    "123",
	Kind:   *groupVersionKind,
	Object: runtime.RawExtension{Raw: data},
}
var arResponseFailure = &admissionv1.AdmissionResponse{
	UID:     arRequest.UID,
	Allowed: false,
	Result:  status,
}
var patchType = admissionv1.PatchType(admissionv1.PatchTypeJSONPatch)
var arResponseMutatingSuccess = &admissionv1.AdmissionResponse{
	UID:       arRequest.UID,
	Allowed:   true,
	PatchType: &patchType,
	Patch:     dataPatch,
}
