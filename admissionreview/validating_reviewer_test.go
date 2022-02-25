package admissionreview_test

import (
	"crypto/rand"
	"testing"

	admissionreview "github.com/ngergs/k8s-adm-ctrl/admissionreview"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidatingReview(t *testing.T) {
	_, err := rand.Read(data)
	assert.Nil(t, err)

	resourceValidatorMock := func(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *admissionreview.ValidateResult {
		assert.Equal(t, groupVersionKind, requestGroupVersionKind)
		assert.Equal(t, data, rawRequest)
		return &admissionreview.ValidateResult{
			Allow:  false,
			Status: status,
		}
	}
	reviewer := admissionreview.ValidatingReviewer(resourceValidatorMock)
	testResult := reviewer.Review(arRequest)
	assert.Equal(t, arResponseFailure, testResult)
}
