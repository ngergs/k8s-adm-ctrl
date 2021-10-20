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

func (*NamespaceLabelMutater) Patch(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *admissionreview.PatchResult {
	if !admissionreview.Contains(compatibleGroupVersionKinds, *requestGroupVersionKind) {
		return &admissionreview.PatchResult{
			Allow: true,
		}
	}
	request := corev1.Namespace{}
	if err := json.Unmarshal(rawRequest, &request); err != nil {
		return &admissionreview.PatchResult{
			Allow:  false,
			Status: admissionreview.GetErrorStatus(http.StatusUnprocessableEntity, "failed to unmarshal into namespace object", err),
		}
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

	return &admissionreview.PatchResult{
		Allow:    true,
		Request:  request,
		Response: response,
	}
}

func (*NamespaceLabelMutater) Validate(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *admissionreview.ValidateResult {
	if !admissionreview.Contains(compatibleGroupVersionKinds, *requestGroupVersionKind) {
		return &admissionreview.ValidateResult{
			Allow: true,
		}
	}
	request := corev1.Namespace{}
	if err := json.Unmarshal(rawRequest, &request); err != nil {
		return &admissionreview.ValidateResult{
			Allow:  false,
			Status: admissionreview.GetErrorStatus(http.StatusUnprocessableEntity, "failed to unmarshal into namespace object", err),
		}
	}

	if request.Labels == nil {
		return &validateFailureResponse
	}
	_, allow := request.Labels[namespaceNameLabelKey]
	var status *metav1.Status = nil
	if !allow {
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
