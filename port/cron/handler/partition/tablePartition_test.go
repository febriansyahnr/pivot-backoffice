package partition

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	partitionTableExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/tablePartitionExt"
	"github.com/stretchr/testify/mock"
)

func TestCreateAccountTransactionPartition(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger)
	}{
		{
			name: "When partition creation was success then should print log success",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)

				loggerMock.On("Info", mock.Anything, "account_transactions partition creation succeeded")
			},
		},
		{
			name: "When partition creation was failed then should print log error",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(errors.New("database down"))

				loggerMock.On("Fatal", mock.Anything, "err-create-partition--account_transactions:", mock.Anything)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := partitionTableExtMock.NewIPartitionTable(t)
			mockLog := loggerMocks.NewILogger(t)
			ctx := context.Background()
			tc.mockSetup(mockSvc, mockLog)

			svc := New(mockLog, mockSvc, nil)
			svc.CreateAccountTransactionPartition(ctx)

			mockSvc.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}

func TestCreateCallbackLogPartition(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger)
	}{
		{
			name: "When partition creation was success then should print log success",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)

				loggerMock.On("Info", mock.Anything, "callback_logs partition creation succeeded")
			},
		},
		{
			name: "When partition creation was failed then should print log error",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(errors.New("database down"))

				loggerMock.On("Fatal", mock.Anything, "err-create-partition--callback_logs:", mock.Anything)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := partitionTableExtMock.NewIPartitionTable(t)
			mockLog := loggerMocks.NewILogger(t)
			ctx := context.Background()
			tc.mockSetup(mockSvc, mockLog)

			svc := New(mockLog, mockSvc, nil)
			svc.CreateCallbackLogPartition(ctx)

			mockSvc.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}

func TestCreatePaymentPartition(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger)
	}{
		{
			name: "When partition creation was success then should print log success",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)

				loggerMock.On("Info", mock.Anything, "payments partition creation succeeded")
			},
		},
		{
			name: "When partition creation was failed then should print log error",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(errors.New("database down"))

				loggerMock.On("Fatal", mock.Anything, "err-create-partition--payments:", mock.Anything)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := partitionTableExtMock.NewIPartitionTable(t)
			mockLog := loggerMocks.NewILogger(t)
			ctx := context.Background()
			tc.mockSetup(mockSvc, mockLog)

			svc := New(mockLog, mockSvc, nil)
			svc.CreatePaymentPartition(ctx)

			mockSvc.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}

func TestCreateDisbursementPartition(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger)
	}{
		{
			name: "When partition creation was success then should print log success",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)

				loggerMock.On("Info", mock.Anything, "disbursements partition creation succeeded")
			},
		},
		{
			name: "When partition creation was failed then should print log error",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(errors.New("database down"))

				loggerMock.On("Fatal", mock.Anything, "err-create-partition--disbursements:", mock.Anything)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := partitionTableExtMock.NewIPartitionTable(t)
			mockLog := loggerMocks.NewILogger(t)
			ctx := context.Background()
			tc.mockSetup(mockSvc, mockLog)

			svc := New(mockLog, mockSvc, nil)
			svc.CreateDisbursementPartition(ctx)

			mockSvc.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}

func TestCreateActivityLogPartition(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger)
	}{
		{
			name: "When partition creation was success then should print log success",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)

				loggerMock.On("Info", mock.Anything, "activity_logs partition creation succeeded")
			},
		},
		{
			name: "When partition creation was failed then should print log error",
			mockSetup: func(mockSvc *partitionTableExtMock.IPartitionTable, loggerMock *loggerMocks.ILogger) {
				mockSvc.On(
					"CreateDayRangePartition",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(errors.New("database down"))

				loggerMock.On("Fatal", mock.Anything, "err-create-partition--activity_logs:", mock.Anything)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := partitionTableExtMock.NewIPartitionTable(t)
			mockLog := loggerMocks.NewILogger(t)
			ctx := context.Background()
			tc.mockSetup(mockSvc, mockLog)

			svc := New(mockLog, mockSvc, nil)
			svc.CreateActivityLogPartition(ctx)

			mockSvc.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}
