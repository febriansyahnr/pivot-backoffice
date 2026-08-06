package callbackService_test

import (
	"testing"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindCallbackByMerchantIdAndCallbackName(t *testing.T) {
	repo := repoMocks.NewICallbackRepository(t)

	var (
		callbackName = "PAYOUT"
		merchantId   = uuid.MustParse("c30998ed-a569-4285-9954-c5815cf6f37e")
	)

	service := New(nil, nil, repo, nil, nil)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *model.Callback
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				repo.On("FindCallbackByNameAndMerchantID", mock.Anything, callbackName, merchantId).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On("FindCallbackByNameAndMerchantID", mock.Anything, callbackName, merchantId).Once().Return(nil, nil)
			},
			wantError: nil, wantResult: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.FindCallbackByMerchantIdAndCallbackName(t.Context(), merchantId, callbackName)

			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
