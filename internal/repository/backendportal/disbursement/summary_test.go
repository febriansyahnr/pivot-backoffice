package disbursementRepository

import (
	"context"
	"testing"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursementDashboard"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSummaryAllToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.GetSummaryAll(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummarySuccessToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.GetSummarySuccess(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryFailedToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.GetSummaryFailed(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryInProgressToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.GetSummaryInProgress(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryWaitingToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummaryWaitingToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummarySingleWaitingToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummarySingleWaitingToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryBulkWaitingToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummaryBulkWaitingToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryWaitingForTopUpToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummaryWaitingForTopUpToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummarySingleWaitingForTopUpToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummarySingleWaitingForTopUpToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryBulkWaitingForTopUpToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummaryBulkWaitingForTopUpToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryRejectedForTopUpToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.GetSummaryRejected(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryApprovedToday(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.GetSummaryApproved(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummarySuccessByBulkID(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummarySuccessByBulkID(ctx, uuid.NewString())

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryFailedByBulkID(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummaryFailedByBulkID(ctx, uuid.NewString())

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryCancelledByBulkID(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummaryCancelledByBulkID(ctx, uuid.NewString())

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestSummaryPendingByBulkID(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.SummaryPendingByBulkID(ctx, uuid.NewString())

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetSummaryByReasonType(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]disbursementDashboardModel.SummaryTransactionByReasonType"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]disbursementDashboardModel.SummaryTransactionByReasonType"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.GetSummaryByReasonType(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			}, constant.StatusFailed)

			assert.NoError(t, err)

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetSummaryByDisbursementStatus(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			repo.GetSummaryByDisbursementStatus(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			}, constant.DisbursementStatusWaiting)

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetActionTransactionSummary(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	merchantId := "4f2a3805-73ab-49f4-97ea-446ee153625f"
	disbursementIds := []string{"0ef16ffd-5905-4d11-bcf3-163b87eba181"}
	db.On("Rebind", constant.StringMockType()).Return("SELECT ???")

	result := disbursementModel.ActionTransactionSummary{
		Total:       12,
		TotalAmount: 1_250_250,
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *disbursementModel.ActionTransactionSummary
	}{
		{
			name: "ERROR:Build query statement",
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), merchantId, disbursementIds,
				).Once().Return("", nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), constant.StringMockType(), mock.Anything,
				).Return("SELECT ???", []interface{}{merchantId, disbursementIds[0]}, nil)

				db.On(
					"GetContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), merchantId, disbursementIds[0],
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr:    constant.ErrSomeErrorForUnitTest,
			wantResult: &disbursementModel.ActionTransactionSummary{},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), merchantId, disbursementIds[0],
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*disbursementModel.ActionTransactionSummary)) = result
				}).Return(nil)
			},
			wantResult: &result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetActionTransactionSummary(context.Background(), merchantId, disbursementIds)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
