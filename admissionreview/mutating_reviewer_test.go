package admissionreview_test

import (
	"encoding/json"
	"testing"

	admissionreview "github.com/ngergs/k8s-adm-ctrl/admissionreview"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMutatingReviewNotAllowed(t *testing.T) {
	resourceMutaterMock := func(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) (result *admissionreview.ValidateResult, patch *admissionreview.Patch) {
		assert.Equal(t, groupVersionKind, requestGroupVersionKind)
		assert.Equal(t, data, rawRequest)
		return &admissionreview.ValidateResult{
			Allow:  false,
			Status: status,
		}, nil
	}

	reviewer := admissionreview.MutatingReviewer(resourceMutaterMock)
	testResult := reviewer.Review(arRequest)
	assert.Equal(t, arResponseFailure, testResult)
}

func TestMutatingReviewAllowed(t *testing.T) {
	var dataUnmarshalled map[string]string
	var dataMutatedUnmarshalled map[string]string
	err := json.Unmarshal(data, &dataUnmarshalled)
	assert.Nil(t, err)
	err = json.Unmarshal(dataMutated, &dataMutatedUnmarshalled)
	assert.Nil(t, err)

	resourceMutaterMock := func(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) (result *admissionreview.ValidateResult, patch *admissionreview.Patch) {
		assert.Equal(t, groupVersionKind, requestGroupVersionKind)
		assert.Equal(t, data, rawRequest)
		return &admissionreview.ValidateResult{
				Allow: true,
			}, &admissionreview.Patch{
				Request:  dataUnmarshalled,
				Response: dataMutatedUnmarshalled,
			}
	}

	reviewer := admissionreview.MutatingReviewer(resourceMutaterMock)
	testResult := reviewer.Review(arRequest)
	assert.Equal(t, arResponseMutatingSuccess, testResult)
}
