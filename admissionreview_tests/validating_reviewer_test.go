package admissionreview_tests

import (
	"crypto/rand"
	"testing"

	"github.com/golang/mock/gomock"
	admissionreview "github.com/selfenergy/k8s-admission-ctrl/admissionreview"
	"github.com/stretchr/testify/assert"
)

func TestValidatingReview(t *testing.T) {
	_, err := rand.Read(data)
	assert.Nil(t, err)

	ctrl := gomock.NewController(t)
	resourceValidatorMock := NewMockResourceValidator(ctrl)
	resourceValidatorMock.
		EXPECT().
		Validate(gomock.Eq(groupVersionKind), gomock.Eq(data)).
		Return(&admissionreview.ValidateResult{
			Allow:  false,
			Status: status,
		})

	reviewer := admissionreview.ValidatingReviewer{
		Validator: resourceValidatorMock,
	}
	testResult := reviewer.Review(arRequest)
	assert.Equal(t, arResponseFailure, testResult)
}
