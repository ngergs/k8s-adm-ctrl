package admissionreview_test

import (
	"encoding/json"
	"testing"

	admissionreview "github.com/ngergs/k8s-adm-ctrl/admissionreview"
	"github.com/stretchr/testify/assert"
)

func TestMutatingReviewNotAllowed(t *testing.T) {
	resourceMutaterMock := func(request *dataType) (*admissionreview.ValidateResult, *admissionreview.Patch[dataType]) {
		//assert.Equal(t, *&arRequest.Object, *reqrequest)
		return &admissionreview.ValidateResult{
			Allow:  false,
			Status: status,
		}, nil
	}

	reviewer := admissionreview.MutatingReviewer(resourceMutaterMock, groupVersionKind)
	testResult := reviewer.Review(arRequest)
	assert.Equal(t, arResponseFailure, testResult)
}

func TestMutatingReviewAllowed(t *testing.T) {
	var dataUnmarshalled dataType
	var dataMutatedUnmarshalled dataType
	err := json.Unmarshal(data, &dataUnmarshalled)
	assert.Nil(t, err)
	err = json.Unmarshal(dataMutated, &dataMutatedUnmarshalled)
	assert.Nil(t, err)

	resourceMutaterMock := func(request *dataType) (*admissionreview.ValidateResult, *admissionreview.Patch[dataType]) {
		//assert.Equal(t, *arRequest, *reqrequest)
		return &admissionreview.ValidateResult{
				Allow: true,
			}, &admissionreview.Patch[dataType]{
				Request:  &dataUnmarshalled,
				Response: &dataMutatedUnmarshalled,
			}
	}

	reviewer := admissionreview.MutatingReviewer(resourceMutaterMock, groupVersionKind)
	testResult := reviewer.Review(arRequest)
	assert.Equal(t, arResponseMutatingSuccess, testResult)
}
