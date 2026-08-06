package merchantTopUp

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindByReferenceNumber(t *testing.T) {
	expectedRef := &merchantTopUp.MerchantTopUp{
		ID:              "uuid-uuid-uuid",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		ReferenceNumber: "reference-number",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	testCases := []struct {
		name           string
		isSuccess      bool
		reference      string
		expectedResult *merchantTopUp.MerchantTopUpResponse
		expectedError  string
		mockSetup      func(mockRepo *repoMocks.IMerchantTopUpRepository)
	}{
		{
			name:           "SUCCESS:Find by reference number",
			isSuccess:      true,
			reference:      "reference-number",
			expectedResult: expectedRef.ToResponse(),
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				mockRepo.On("GetByReferenceNumber", mock.Anything, mock.Anything).Return(expectedRef, nil)
			},
		},
		{
			name:          "ERROR:Data not found",
			isSuccess:     false,
			reference:     "not-found",
			expectedError: "merchant top up reference not found",
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				mockRepo.On("GetByReferenceNumber", mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		{
			name:          "ERROR:Error find user",
			isSuccess:     false,
			reference:     "user-error",
			expectedError: "some error",
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				mockRepo.On("GetByReferenceNumber", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	cfg := &config.Config{
		ServiceName: "testing",
	}
	buf := new(bytes.Buffer)
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			refRepo := repoMocks.NewIMerchantTopUpRepository(t)
			snapCoreMock := repoMocks.NewISnapCoreRepository(t)

			tc.mockSetup(refRepo)

			trxSvc := New(cfg, log, nil, refRepo, snapCoreMock)

			response, err := trxSvc.FindByReferenceNumber(context.Background(), tc.reference)
			if tc.isSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			} else {
				require.Error(t, err)
				require.Empty(t, response)
				require.True(t, strings.Contains(err.Error(), tc.expectedError))
			}

			refRepo.AssertExpectations(t)
			snapCoreMock.AssertExpectations(t)
		})
	}
}

func TestFindByMerchantAccountNameAndPaymentMethodId(t *testing.T) {
	expectedRef := &merchantTopUp.MerchantTopUp{
		ID:              "uuid-uuid-uuid",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		ReferenceNumber: "reference-number",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	testCases := []struct {
		name           string
		isSuccess      bool
		paymentMethod  string
		merchantID     string
		expectedResult *merchantTopUp.MerchantTopUpResponse
		expectedError  string
		mockSetup      func(mockRepo *repoMocks.IMerchantTopUpRepository)
	}{
		{
			name:           "SUCCESS",
			isSuccess:      true,
			paymentMethod:  "payment-method-id",
			merchantID:     "merchant-id",
			expectedResult: expectedRef.ToResponse(),
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				mockRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(expectedRef, nil)
			},
		},
		{
			name:          "ERROR:Data not found",
			isSuccess:     false,
			paymentMethod: "payment-method-id",
			merchantID:    "not-found",
			expectedError: "is not found",
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				mockRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil, nil)
			},
		},
		{
			name:          "ERROR:Some error",
			isSuccess:     false,
			paymentMethod: "payment-method-id",
			merchantID:    "error-find",
			expectedError: "some error",
			mockSetup: func(mockRepo *repoMocks.IMerchantTopUpRepository) {
				mockRepo.On(
					"GetByMerchantAccountNameAndPaymentMethodId", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	cfg := &config.Config{
		ServiceName: "testing",
	}
	buf := new(bytes.Buffer)
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repoMocks.NewIMerchantTopUpRepository(t)

			tc.mockSetup(repo)

			trxSvc := New(cfg, log, nil, repo, nil)

			response, err := trxSvc.FindByMerchantAccountNameAndPaymentMethodId(context.Background(), tc.merchantID, constant.TypeDisbursement, tc.paymentMethod)
			if tc.isSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)

			} else {
				require.Error(t, err)
				require.Empty(t, response)
				require.True(t, strings.Contains(err.Error(), tc.expectedError))
			}
			repo.AssertExpectations(t)
		})
	}
}
