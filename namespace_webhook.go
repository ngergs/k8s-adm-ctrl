package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/selfenergy/k8s-admission-ctrl/admissionreview"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const namespaceNameLabelKey = "kubernetes.io/metadata.name"

var compatibleGroupVersionKinds = []metav1.GroupVersionKind{{
	Group:   "",
	Version: "v1",
	Kind:    "Namespace",
}}

var validateFailureResponse = admissionreview.ValidateResult{
	Allow: false,
	Status: &metav1.Status{
		Status:  "Failure",
		Message: fmt.Sprintf("The label %v has to be mandatory set.", namespaceNameLabelKey),
		Code:    http.StatusUnprocessableEntity,
	},
}

type NamespaceLabelMutater struct{}

// unmarshallToNamespace checks if the requestGroupVersionKind fits to a namespace object and unmarshalls the raw request into a namespace struct if this is the case.
// The presence of the validateResult implies that the skip condition has been fulfilled or an error occured during unmarshalling. namespace is nil then
func unmarshallToNamespace(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) (*corev1.Namespace, *admissionreview.ValidateResult) {
	if !admissionreview.Contains(compatibleGroupVersionKinds, *requestGroupVersionKind) {
		return nil, &admissionreview.ValidateResult{
			Allow: true,
		}
	}
	namespace := corev1.Namespace{}
	if err := json.Unmarshal(rawRequest, &namespace); err != nil {
		return nil, &admissionreview.ValidateResult{
			Allow:  false,
			Status: admissionreview.GetErrorStatus(http.StatusUnprocessableEntity, "failed to unmarshal into namespace object", err),
		}
	}
	return &namespace, nil
}

func (*NamespaceLabelMutater) Patch(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) (*admissionreview.ValidateResult, *admissionreview.Patches) {
	request, errorValidateResult := unmarshallToNamespace(requestGroupVersionKind, rawRequest)
	if errorValidateResult != nil {
		return errorValidateResult, nil
	}

	// copy structure to make changes
	var response = request.DeepCopy()
	if response.Labels == nil {
		response.Labels = make(map[string]string)
	}
	if _, ok := response.Labels[namespaceNameLabelKey]; !ok {
		log.Info().Msgf("For namespace %v the %v label is missing, it has been added.", response.Name, namespaceNameLabelKey)
		response.Labels[namespaceNameLabelKey] = response.Name
	}

	return &admissionreview.ValidateResult{
			Allow: true,
		}, &admissionreview.Patches{
			Request:  request,
			Response: response,
		}
}

func (*NamespaceLabelMutater) Validate(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *admissionreview.ValidateResult {
	request, errorValidateResult := unmarshallToNamespace(requestGroupVersionKind, rawRequest)
	if errorValidateResult != nil {
		return errorValidateResult
	}

	if request.Labels == nil {
		return &validateFailureResponse
	}
	_, allow := request.Labels[namespaceNameLabelKey]
	var status *metav1.Status = nil
	if !allow {
		log.Info().Msgf("Request for namespace %v failed validation. The label %v is missing.", request.Name, namespaceNameLabelKey)
		status = &metav1.Status{
			Status:  "Failure",
			Message: fmt.Sprintf("The label %v has to be mandatory set.", namespaceNameLabelKey),
			Code:    http.StatusUnprocessableEntity,
		}
	}
	return &admissionreview.ValidateResult{
		Allow:  allow,
		Status: status,
	}
}
