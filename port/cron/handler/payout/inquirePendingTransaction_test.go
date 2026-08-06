package payout

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/mock"
)

func TestInquirePendingPayout(t *testing.T) {
	mockLogger := loggerMock.NewILogger(t)
	mockDisbursementSvc := serviceMock.NewIDisbursementService(t)

	handler := &cronHandler{
		service: mockDisbursementSvc,
		log:     mockLogger,
		config: &config.Config{
			DisbursementConfig: config.DisbursementConfig{
				RetryInquiryConfig: config.DisbursementRetryInquiryConfig{
					DelayTimeMinute:     0,
					RetryIntervalMinute: 5,
				},
			},
		},
	}

	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "success",
			setupMock: func() {
				mockDisbursementSvc.On("RetryInquirePendingTransactions", constant.ValueCtxMockType(), mock.Anything, mock.Anything).Return(&disbursementModel.RetryInquireDisbuesementSummary{
					Total:           1,
					Amount:          10_000,
					TotalSucceeded:  1,
					AmountSucceeded: 10_000,
				}, nil).Once()
				mockLogger.On("Info", constant.ValueCtxMockType(), "Retry to inquire pending transactions", mock.Anything, mock.Anything).Once()
				mockLogger.On("Info", constant.ValueCtxMockType(), "Inquiry pending transactions completed", mock.Anything).Once()
			},
		},
		{
			name: "success with failed to retry",
			setupMock: func() {
				mockDisbursementSvc.On("RetryInquirePendingTransactions", constant.ValueCtxMockType(), mock.Anything, mock.Anything).Return(&disbursementModel.RetryInquireDisbuesementSummary{
					Total:           2,
					Amount:          12_000,
					TotalSucceeded:  1,
					AmountSucceeded: 10_000,
					TotalFailed:     1,
					AmountFailed:    2000,
				}, nil).Once()
				mockLogger.On("Info", constant.ValueCtxMockType(), "Retry to inquire pending transactions", mock.Anything, mock.Anything).Once()
				mockLogger.On("Info", constant.ValueCtxMockType(), "Inquiry pending transactions completed", mock.Anything).Once()
				mockLogger.On("Warn", constant.ValueCtxMockType(), "Some transactions failed to retry", mock.Anything, mock.Anything).Once()
			},
		},
		{
			name: "failure",
			setupMock: func() {
				mockDisbursementSvc.On("RetryInquirePendingTransactions", constant.ValueCtxMockType(), mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
				mockLogger.On("Info", constant.ValueCtxMockType(), "Retry to inquire pending transactions", mock.Anything, mock.Anything).Once()
				mockLogger.On("Fatal", constant.ValueCtxMockType(), "Failed to inquire pending transactions", mock.Anything).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			handler.InquirePendingPayout(ctx)

			mockDisbursementSvc.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
