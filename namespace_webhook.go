package main

import (
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

type NamespaceLabelMutater struct{}

func (*NamespaceLabelMutater) Patch(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) (*admissionreview.ValidateResult, *admissionreview.Patches) {
	var request corev1.Namespace
	validateSkipErrorResult := admissionreview.UnmarshallAdmissionRequest(compatibleGroupVersionKinds, &request, requestGroupVersionKind, rawRequest)
	if validateSkipErrorResult != nil {
		return validateSkipErrorResult, nil
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

func isNamespaceLabelPresent(labels map[string]string) bool {
	if labels == nil {
		return false
	}
	_, ok := labels[namespaceNameLabelKey]
	return ok
}

func (*NamespaceLabelMutater) Validate(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *admissionreview.ValidateResult {
	var request corev1.Namespace
	validateSkipErrorResult := admissionreview.UnmarshallAdmissionRequest(compatibleGroupVersionKinds, &request, requestGroupVersionKind, rawRequest)
	if validateSkipErrorResult != nil {
		return validateSkipErrorResult
	}

	if !isNamespaceLabelPresent(request.Labels) {
		log.Info().Msgf("Request for namespace %v failed validation. The label %v is missing.", request.Name, namespaceNameLabelKey)
		return &admissionreview.ValidateResult{
			Allow: false,
			Status: &metav1.Status{
				Status:  "Failure",
				Message: fmt.Sprintf("The label %v is absent, but has to be mandatory set.", namespaceNameLabelKey),
				Code:    http.StatusUnprocessableEntity,
			},
		}
	}

	log.Info().Msgf("Request for namespace %v passed validation. The label %v is present.", request.Name, namespaceNameLabelKey)
	return &admissionreview.ValidateResult{
		Allow: true,
	}
}
