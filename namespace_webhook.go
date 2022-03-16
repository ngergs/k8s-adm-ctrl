package main

import (
	"fmt"
	"net/http"

	"github.com/ngergs/k8s-adm-ctrl/admissionreview"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const namespaceNameLabelKey = "kubernetes.io/metadata.name"

var compatibleGroupVersionKind = &metav1.GroupVersionKind{
	Group:   "",
	Version: "v1",
	Kind:    "Namespace",
}

var validationAllow = &admissionreview.ValidateResult{
	Allow: true,
}
var labelAbsendValidationError = &admissionreview.ValidateResult{
	Allow: false,
	Status: &metav1.Status{
		Status:  "Failure",
		Message: fmt.Sprintf("The label %v is absent, but has to be mandatory set.", namespaceNameLabelKey),
		Code:    http.StatusUnprocessableEntity,
	},
}

// NamespaceLabelMutater is an example struct that implements the admissionreview.ResourceMutater and admissionreview.ResourceValidator interfaces
// to add a namespaceLabelKey if absent.
type NamespaceLabelMutater struct{}

// Patch implements the admissionreview.ResourceMutater interface and serves as an example implementation to add a namespaceLabelKey if absent.
func (*NamespaceLabelMutater) Patch(request *corev1.Namespace) (*admissionreview.ValidateResult, *admissionreview.Patch[corev1.Namespace]) {
	if _, ok := request.Labels[namespaceNameLabelKey]; ok {
		log.Info().Msgf("For namespace %v the %v label is present, no mutation applied.", request.Name, namespaceNameLabelKey)
		return validationAllow, nil
	}

	// copy structure to make changes for JSON diff later on
	var response = request.DeepCopy()
	if response.Labels == nil {
		response.Labels = make(map[string]string)
	}
	response.Labels[namespaceNameLabelKey] = response.Name
	log.Info().Msgf("For namespace %v the %v label is missing, it has been added.", response.Name, namespaceNameLabelKey)

	patch := &admissionreview.Patch[corev1.Namespace]{
		Request:  request,
		Response: response,
	}
	return validationAllow, patch
}

// Validate implements the admissionreview.ResourceValidator interface and serves as an example implementation to check whethera namespaceLabelKey is present.
func (*NamespaceLabelMutater) Validate(request *corev1.Namespace) *admissionreview.ValidateResult {
	if _, ok := request.Labels[namespaceNameLabelKey]; !ok {
		log.Info().Msgf("Request for namespace %v failed validation. The label %v is missing.", request.Name, namespaceNameLabelKey)
		return labelAbsendValidationError
	}

	log.Info().Msgf("Request for namespace %v passed validation. The label %v is present.", request.Name, namespaceNameLabelKey)
	return &admissionreview.ValidateResult{
		Allow: true,
	}
}
