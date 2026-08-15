package disbursementRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testCaseMySQLError = "ERROR: Mysql error"
)

func TestCountByIDsAndMerchantID(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		inputIDs  []string
		count     int
	}{
		{
			name:     "SUCCESS",
			inputIDs: []string{uuid.NewString(), uuid.NewString()},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"In",
					mock.AnythingOfType("string"), constant.ArrayStringMockType(), mock.AnythingOfType("string"),
				).Return("", []interface{}{}, nil)
				mysqlMock.On("Rebind", mock.AnythingOfType("string")).Return("")
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(nil)
			},
		},
		{
			name:     testCaseMySQLError,
			inputIDs: []string{uuid.NewString()},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"In",
					mock.AnythingOfType("string"), constant.ArrayStringMockType(), mock.AnythingOfType("string"),
				).Return("", []interface{}{}, nil)
				mysqlMock.On("Rebind", mock.AnythingOfType("string")).Return("")
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
		},
		{
			name:     "ERROR:Build query",
			inputIDs: []string{uuid.NewString()},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"In",
					mock.AnythingOfType("string"), constant.ArrayStringMockType(), mock.AnythingOfType("string"),
				).Return("", nil, errors.New("invalid arguments"))

			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			count := repo.CountByIDsAndMerchantID(ctx, tc.inputIDs, uuid.NewString())
			assert.Equal(t, count, tc.count)

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCountByBulkID(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		count     int
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			count: 0,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			count: 0,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			count := repo.CountByBulkID(ctx, uuid.NewString())
			assert.Equal(t, count, tc.count)

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCountWaitingSingleDisbursement(t *testing.T) {
	testCases := []struct {
		name           string
		mockSetup      func(mysqlMock *mysqlMocks.IMySqlExt)
		filter         disbursementDashboardModel.GetDisbursementDashboardFilter
		expectedTotals disbursementDashboardModel.SummaryTransactionDTO
		expectError    bool
	}{
		{
			name: "SUCCESS",
			filter: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			expectedTotals: disbursementDashboardModel.SummaryTransactionDTO{Count: 0, Sum: 0},
			expectError:    false,
		},
		{
			name: "SUCCESS: IsXbPayout true",
			filter: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: true,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			expectedTotals: disbursementDashboardModel.SummaryTransactionDTO{Count: 0, Sum: 0},
			expectError:    false,
		},
		{
			name: testCaseMySQLError,
			filter: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("mysql error"))

			},
			expectedTotals: disbursementDashboardModel.SummaryTransactionDTO{},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			totals, err := repo.CountWaitingSingleDisbursement(ctx, tc.filter)
			assert.Equal(t, tc.expectedTotals, totals)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
func TestCountWaitingBulkDisbursement(t *testing.T) {
	testCases := []struct {
		name           string
		mockSetup      func(mysqlMock *mysqlMocks.IMySqlExt)
		expectedTotals disbursementDashboardModel.SummaryTransactionDTO
		expectError    bool
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
					constant.StringMockType(),
				).Return(nil)
			},
			expectedTotals: disbursementDashboardModel.SummaryTransactionDTO{Count: 0, Sum: 0},
			expectError:    false,
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
					constant.StringMockType(),
				).Return(errors.New("mysql error"))

			},
			expectedTotals: disbursementDashboardModel.SummaryTransactionDTO{},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			totals, err := repo.CountWaitingBulkDisbursement(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})
			assert.Equal(t, tc.expectedTotals, totals)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCountPendingSingleDisbursement(t *testing.T) {
	testCases := []struct {
		name          string
		mockSetup     func(mysqlMock *mysqlMocks.IMySqlExt)
		filter        disbursementDashboardModel.GetDisbursementDashboardFilter
		expectedTotal disbursementDashboardModel.SummaryTransactionDTO
		expectError   bool
	}{
		{
			name: "SUCCESS",
			filter: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {

				expectedTotals := disbursementDashboardModel.SummaryTransactionDTO{
					Count: 5,
					Sum:   1500.00,
				}

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*disbursementDashboardModel.SummaryTransactionDTO)
					*arg = expectedTotals
				}).Once()
			},
			expectedTotal: disbursementDashboardModel.SummaryTransactionDTO{
				Count: 5,
				Sum:   1500.00,
			},
			expectError: false,
		},
		{
			name: "SUCCESS: IsXbPayout true",
			filter: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: true,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {

				expectedTotals := disbursementDashboardModel.SummaryTransactionDTO{
					Count: 3,
					Sum:   900.00,
				}

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*disbursementDashboardModel.SummaryTransactionDTO)
					*arg = expectedTotals
				}).Once()
			},
			expectedTotal: disbursementDashboardModel.SummaryTransactionDTO{
				Count: 3,
				Sum:   900.00,
			},
			expectError: false,
		},
		{
			name: "MYSQL ERROR",
			filter: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("mysql error")).Once()

			},
			expectedTotal: disbursementDashboardModel.SummaryTransactionDTO{},
			expectError:   true,
		},
		{
			name: "NO RESULTS",
			filter: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				expectedTotals := disbursementDashboardModel.SummaryTransactionDTO{
					Count: 0,
					Sum:   0.00,
				}

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*disbursementDashboardModel.SummaryTransactionDTO"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*disbursementDashboardModel.SummaryTransactionDTO)
					*arg = expectedTotals
				}).Once()
			},
			expectedTotal: disbursementDashboardModel.SummaryTransactionDTO{
				Count: 0,
				Sum:   0.00,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "disbursements")

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			totals, err := repo.CountPendingSingleDisbursement(ctx, tc.filter)

			assert.Equal(t, tc.expectedTotal, totals)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCountPendingBulkDisbursement(t *testing.T) {
	testCases := []struct {
		name           string
		mockSetup      func(mysqlMock *mysqlMocks.IMySqlExt)
		expectedTotals disbursementDashboardModel.SummaryTransactionDTO
		expectError    bool
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
					constant.StringMockType(),
				).Return(nil)
			},
			expectedTotals: disbursementDashboardModel.SummaryTransactionDTO{Count: 0, Sum: 0},
			expectError:    false,
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
					constant.StringMockType(),
				).Return(errors.New("mysql error"))

			},
			expectedTotals: disbursementDashboardModel.SummaryTransactionDTO{},
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			totals, err := repo.CountPendingBulkDisbursement(ctx, disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: uuid.NewString(),
				IsXbPayout: false,
			})
			assert.Equal(t, tc.expectedTotals, totals)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCountStatusInProgressByBulkID(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		count     int
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			count: 0,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			count: 0,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			count := repo.CountStatusInProgressByBulkID(ctx, uuid.NewString())
			assert.Equal(t, count, tc.count)
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCountByMerchantAndReference(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		count     int
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			count: 0,
		},
		{
			name: testCaseMySQLError,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			count: 0,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			count := repo.CountByMerchantAndReference(ctx, uuid.NewString(), "reference")
			assert.Equal(t, count, tc.count)

			mockMysql.AssertExpectations(t)

		})
	}
}
