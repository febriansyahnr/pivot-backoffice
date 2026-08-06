package refundProcessorService

import (
	"context"
	"errors"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	snapQReModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQRISStrategy_Process(t *testing.T) {

	_, pdkLogger, err := test.SetupLogger()
	assert.NoError(t, err)

	// Define test cases
	testCases := []struct {
		name           string
		request        *refundModel.RefundProcessRequest
		setupMocks     func(snapCoreRepo *mocks.ISnapCoreRepository)
		expectedError  error
		expectLogError bool
		expectLogInfo  bool
	}{
		{
			name: "when the refund was success, then should not return error",
			request: &refundModel.RefundProcessRequest{
				RefundID:           "refund-123",
				PaymentProcessorID: "qr-123",
				Refund: &refundModel.Refund{
					Reason:   "customer requested",
					Currency: "IDR",
					Amount:   10000.0,
				},
			},
			setupMocks: func(snapCoreRepo *mocks.ISnapCoreRepository) {
				snapCoreRepo.On("RefundQRMPM", mock.Anything, &snapQReModel.QRMPMRefundRequest{
					QRID:   "qr-123",
					Reason: "customer requested",
					AdditionalInfo: map[string]interface{}{
						"clientReferenceId": "",
					},
					Amount: commonModel.Amount{
						Currency: "IDR",
						Value:    "10000.00",
					},
				}).Return(&snapQReModel.RefundResponseData{
					RefundNo:        "refund-123",
					ResponseCode:    "200",
					ResponseMessage: "Success",
				}, nil)
			},
			expectedError: nil,
		},
		{
			name: "processor returns error",
			request: &refundModel.RefundProcessRequest{
				RefundID:           "refund-123",
				PaymentProcessorID: "qr-123",
				Refund: &refundModel.Refund{
					Reason:   "customer requested",
					Currency: "IDR",
					Amount:   10000.0,
				},
			},
			setupMocks: func(snapCoreRepo *mocks.ISnapCoreRepository) {
				snapCoreRepo.On("RefundQRMPM", mock.Anything, &snapQReModel.QRMPMRefundRequest{
					QRID:   "qr-123",
					Reason: "customer requested",
					AdditionalInfo: map[string]interface{}{
						"clientReferenceId": "",
					},
					Amount: commonModel.Amount{
						Currency: "IDR",
						Value:    "10000.00",
					},
				}).Return(nil, errors.New("processor error"))
			},
			expectedError:  errors.New("processor error"),
			expectLogError: true,
			expectLogInfo:  false,
		},
		{
			name: "processor returns nil response",
			request: &refundModel.RefundProcessRequest{
				RefundID:           "refund-123",
				PaymentProcessorID: "qr-123",
				Refund: &refundModel.Refund{
					Reason:   "customer requested",
					Currency: "IDR",
					Amount:   10000.0,
				},
			},
			setupMocks: func(snapCoreRepo *mocks.ISnapCoreRepository) {
				snapCoreRepo.On("RefundQRMPM", mock.Anything, &snapQReModel.QRMPMRefundRequest{
					QRID:   "qr-123",
					Reason: "customer requested",
					AdditionalInfo: map[string]interface{}{
						"clientReferenceId": "",
					},
					Amount: commonModel.Amount{
						Currency: "IDR",
						Value:    "10000.00",
					},
				}).Return(nil, nil)
			},
			expectedError:  constant.ErrFailedToRefundPayment,
			expectLogError: true,
			expectLogInfo:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockSnapCoreRepo := mocks.NewISnapCoreRepository(t)
			tc.setupMocks(mockSnapCoreRepo)

			defer mockSnapCoreRepo.AssertExpectations(t)

			// Create strategy
			strategy := &QRISStrategy{
				logger:       pdkLogger,
				snapCoreRepo: mockSnapCoreRepo,
			}

			// Execute test
			err := strategy.Process(context.Background(), tc.request)

			// Assertions
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
