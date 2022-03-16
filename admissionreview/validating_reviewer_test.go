package admissionreview_test

import (
	"testing"

	admissionreview "github.com/ngergs/k8s-adm-ctrl/admissionreview"
	"github.com/stretchr/testify/assert"
)

func TestValidatingReview(t *testing.T) {
	resourceValidatorMock := func(request *dataType) *admissionreview.ValidateResult {
		return &admissionreview.ValidateResult{
			Allow:  false,
			Status: status,
		}
	}
	reviewer := admissionreview.ValidatingReviewer(resourceValidatorMock, groupVersionKind)
	testResult := reviewer.Review(arRequest)
	assert.Equal(t, arResponseFailure, testResult)
}
