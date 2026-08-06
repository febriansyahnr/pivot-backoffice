package disbursementRepository

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// Test cases for IsXbPayout: true

func TestSummaryAllTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
		"GetContext",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
		constant.StringMockType(),
		constant.StringMockType(),
		mock.AnythingOfType(constant.MockTypeTime),
		mock.AnythingOfType(constant.MockTypeTime),
		constant.StringMockType(),
	).Return(nil)

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.GetSummaryAll(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummarySuccessTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
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

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.GetSummarySuccess(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummaryFailedTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
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

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.GetSummaryFailed(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummaryInProgressTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
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

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.GetSummaryInProgress(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummaryWaitingTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
		"GetContext",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
		constant.StringMockType(),
		constant.StringMockType(),
		mock.AnythingOfType(constant.MockTypeTime),
		constant.StringMockType(),
		constant.StringMockType(),
	).Return(nil)

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.SummaryWaitingToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummarySingleWaitingTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
		"GetContext",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
		constant.StringMockType(),
		constant.StringMockType(),
		mock.AnythingOfType(constant.MockTypeTime),
		constant.StringMockType(),
		constant.StringMockType(),
	).Return(nil)

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.SummarySingleWaitingToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummaryBulkWaitingTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
		"GetContext",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
		constant.StringMockType(),
		constant.StringMockType(),
		mock.AnythingOfType(constant.MockTypeTime),
		constant.StringMockType(),
		constant.StringMockType(),
	).Return(nil)

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.SummaryBulkWaitingToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummaryWaitingForTopUpTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
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

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.SummaryWaitingForTopUpToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummarySingleWaitingForTopUpTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
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

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.SummarySingleWaitingForTopUpToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummaryBulkWaitingForTopUpTodayWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
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

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.SummaryBulkWaitingForTopUpToday(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummaryRejectedWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
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

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.GetSummaryRejected(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}

func TestSummaryApprovedWithXbPayout(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	mockMysql.On(
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

	ctx := context.Background()
	repo := New(mockMysql, mockLogger)
	repo.GetSummaryApproved(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
		MerchantID: uuid.NewString(),
		IsXbPayout: true,
	})

	mockMysql.AssertExpectations(t)
}
