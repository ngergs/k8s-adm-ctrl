package admissionreview_tests

import (
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	admissionreview "github.com/selfenergy/k8s-admission-ctrl/admissionreview"
	"github.com/stretchr/testify/assert"
)

func TestMutatingReviewNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	resourceMutatorMock := NewMockResourceMutater(ctrl)
	resourceMutatorMock.
		EXPECT().
		Patch(gomock.Eq(groupVersionKind), gomock.Eq(data)).
		Return(&admissionreview.ValidateResult{
			Allow:  false,
			Status: status,
		}, nil)

	reviewer := admissionreview.MutatingReviewer{
		Mutater: resourceMutatorMock,
	}
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

	ctrl := gomock.NewController(t)
	resourceMutatorMock := NewMockResourceMutater(ctrl)
	resourceMutatorMock.
		EXPECT().
		Patch(gomock.Eq(groupVersionKind), gomock.Eq(data)).
		Return(&admissionreview.ValidateResult{
			Allow: true,
		}, &admissionreview.Patches{
			Request:  dataUnmarshalled,
			Response: dataMutatedUnmarshalled,
		})

	reviewer := admissionreview.MutatingReviewer{
		Mutater: resourceMutatorMock,
	}
	testResult := reviewer.Review(arRequest)
	assert.Equal(t, arResponseMutatingSuccess, testResult)
}
