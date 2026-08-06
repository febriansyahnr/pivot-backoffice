package refundProcessorService

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/test"
)

func TestCardStrategyProcess(t *testing.T) {

	_, pdkLogger, err := test.SetupLogger()
	assert.NoError(t, err)

	testCases := []struct {
		name          string
		request       *refundModel.RefundProcessRequest
		setupMocks    func(*mocks.ICreditcardCoreProcessorRepository)
		expectedError error
	}{
		{
			name: "success - card refund processed successfully",
			request: &refundModel.RefundProcessRequest{
				RefundID:                 "refund-123",
				PaymentClientReferenceID: "client-ref-123",
				PaymentProcessorID:       "acquirer-123",
				Refund: &refundModel.Refund{
					MerchantID: "merchant-123",
					Reason:     "customer requested",
				},
			},
			setupMocks: func(creditCardRepo *mocks.ICreditcardCoreProcessorRepository) {
				creditCardRepo.On("Refund", mock.Anything, &creditcardCoreProcessorModel.RefundRequest{
					MerchantID:            "merchant-123",
					ClientTransactionID:   "client-ref-123",
					AcquirerTransactionID: "acquirer-123",
				}).Return(&creditcardCoreProcessorModel.RefundResponseData{
					Status: constant.CreditCardStatusSuccess,
				}, nil)
			},
			expectedError: nil,
		},
		{
			name: "fail - credit card repository returns error",
			request: &refundModel.RefundProcessRequest{
				RefundID:                 "refund-123",
				PaymentClientReferenceID: "client-ref-123",
				PaymentProcessorID:       "acquirer-123",
				Refund: &refundModel.Refund{
					MerchantID: "merchant-123",
					Reason:     "customer requested",
				},
			},
			setupMocks: func(creditCardRepo *mocks.ICreditcardCoreProcessorRepository) {
				creditCardRepo.On("Refund", mock.Anything, &creditcardCoreProcessorModel.RefundRequest{
					MerchantID:            "merchant-123",
					ClientTransactionID:   "client-ref-123",
					AcquirerTransactionID: "acquirer-123",
				}).Return(nil, errors.New("repository error"))
			},
			expectedError: errors.New("repository error"),
		},
		{
			name: "fail - credit card repository returns nil response",
			request: &refundModel.RefundProcessRequest{
				RefundID:                 "refund-123",
				PaymentClientReferenceID: "client-ref-123",
				PaymentProcessorID:       "acquirer-123",
				Refund: &refundModel.Refund{
					MerchantID: "merchant-123",
					Reason:     "customer requested",
				},
			},
			setupMocks: func(creditCardRepo *mocks.ICreditcardCoreProcessorRepository) {
				creditCardRepo.On("Refund", mock.Anything, &creditcardCoreProcessorModel.RefundRequest{
					MerchantID:            "merchant-123",
					ClientTransactionID:   "client-ref-123",
					AcquirerTransactionID: "acquirer-123",
				}).Return(nil, nil)
			},
			expectedError: errors.New("ERROR_UNPROCESSABLE_CONTENT | data not found"),
		},
		{
			name: "fail - refund status is not success",
			request: &refundModel.RefundProcessRequest{
				RefundID:                 "refund-123",
				PaymentClientReferenceID: "client-ref-123",
				PaymentProcessorID:       "acquirer-123",
				Refund: &refundModel.Refund{
					MerchantID: "merchant-123",
					Reason:     "customer requested",
				},
			},
			setupMocks: func(creditCardRepo *mocks.ICreditcardCoreProcessorRepository) {
				creditCardRepo.On("Refund", mock.Anything, &creditcardCoreProcessorModel.RefundRequest{
					MerchantID:            "merchant-123",
					ClientTransactionID:   "client-ref-123",
					AcquirerTransactionID: "acquirer-123",
				}).Return(&creditcardCoreProcessorModel.RefundResponseData{
					Status: constant.CreditCardStatusFailed,
				}, nil)
			},
			expectedError: errors.New("ERROR_UNPROCESSABLE_CONTENT | failed to refund payment"),
		},
		{
			name: "fail - refund status is blocked",
			request: &refundModel.RefundProcessRequest{
				RefundID:                 "refund-123",
				PaymentClientReferenceID: "client-ref-123",
				PaymentProcessorID:       "acquirer-123",
				Refund: &refundModel.Refund{
					MerchantID: "merchant-123",
					Reason:     "customer requested",
				},
			},
			setupMocks: func(creditCardRepo *mocks.ICreditcardCoreProcessorRepository) {
				creditCardRepo.On("Refund", mock.Anything, &creditcardCoreProcessorModel.RefundRequest{
					MerchantID:            "merchant-123",
					ClientTransactionID:   "client-ref-123",
					AcquirerTransactionID: "acquirer-123",
				}).Return(&creditcardCoreProcessorModel.RefundResponseData{
					Status: constant.CreditCardStatusBlocked,
				}, nil)
			},
			expectedError: errors.New("ERROR_UNPROCESSABLE_CONTENT | failed to refund payment"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCreditCardRepo := mocks.NewICreditcardCoreProcessorRepository(t)
			tc.setupMocks(mockCreditCardRepo)

			defer mockCreditCardRepo.AssertExpectations(t)

			strategy := &CardStrategy{
				logger:         pdkLogger,
				creditCardRepo: mockCreditCardRepo,
			}

			err := strategy.Process(context.Background(), tc.request)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
