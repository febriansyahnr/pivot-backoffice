package account

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	mockLogger "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/mock"
)

func TestCalculateAllMerchantEODBalance(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mockSvc *serviceMocks.IAccountService, loggerMock *loggerMocks.ILogger)
	}{
		{
			name: "SUCCESS: Calculate All Merchant EOD Balance",
			mockSetup: func(mockSvc *serviceMocks.IAccountService, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CalculateAccountEodBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Return(nil)
			},
		},
		{
			name: "ERROR: CalculateAccountEodBalance",
			mockSetup: func(mockSvc *serviceMocks.IAccountService, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CalculateAccountEodBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Return(constant.ErrSomeErrorForUnitTest)

				loggerMock.On("Fatal", mock.Anything, "error when calculate eod balance", mock.Anything)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := serviceMocks.NewIAccountService(t)
			mockLog := mockLogger.NewILogger(t)
			ctx := context.Background()
			tc.mockSetup(mockSvc, mockLog)

			svc := NewAccount(mockLog, mockSvc)
			svc.CalculateAllMerchantEODBalance(ctx)

			mockSvc.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}

func TestCalculateAllMerchantDailyAccountTransaction(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mockSvc *serviceMocks.IAccountService, loggerMock *loggerMocks.ILogger)
	}{
		{
			name: "SUCCESS: Calculate Daily Account Transaction",
			mockSetup: func(mockSvc *serviceMocks.IAccountService, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CalculateDailyAccountTransaction",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*time.Location"),
				).Return(nil)
			},
		},
		{
			name: "ERROR: CalculateDailyAccountTransaction",
			mockSetup: func(mockSvc *serviceMocks.IAccountService, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CalculateDailyAccountTransaction",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*time.Location"),
				).Return(constant.ErrSomeErrorForUnitTest)

				loggerMock.On("Fatal", mock.Anything, "error when calculate daily account transaction", mock.Anything)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := serviceMocks.NewIAccountService(t)
			mockLog := mockLogger.NewILogger(t)
			ctx := context.Background()
			tc.mockSetup(mockSvc, mockLog)

			svc := NewAccount(mockLog, mockSvc)
			svc.CalculateAllMerchantDailyAccountTransaction(ctx, time.Now().Location())

			mockSvc.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}
