package accounttransaction_repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountTransactionRepositoryGetAggregateTransactions(t *testing.T) {
	aggregateResponse := &orchestrator_model.AggregateResponse{
		SumOfDebit:  0,
		SumOfCredit: 0,
	}
	request := &orchestrator_model.GetAggregateRequest{
		MerchantID: uuid.New(),
		AccountID:  uuid.New(),
		Statuses:   []string{constant.StatusPending, constant.StatusSuccess},
		StartAt:    &util.TimeNow,
		EndAt:      &util.TimeNow,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *orchestrator_model.AggregateResponse
		wantErr   bool
		request   *orchestrator_model.GetAggregateRequest
	}{
		{
			name: "SUCCESS: GetAggregateTransactions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					aggregateResponsePtr := args.Get(1).(*orchestrator_model.AggregateResponse)
					*aggregateResponsePtr = *aggregateResponse
				})
			},
			expected: aggregateResponse,
			wantErr:  false,
		},
		{
			name: "SUCCESS: GetAggregateTransactions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					aggregateResponsePtr := args.Get(1).(*orchestrator_model.AggregateResponse)
					*aggregateResponsePtr = *aggregateResponse
				})
			},
			expected: aggregateResponse,
			wantErr:  false,
			request: &orchestrator_model.GetAggregateRequest{
				MerchantID: uuid.New(),
				AccountIDs: []string{uuid.New().String()},
				Statuses:   []string{constant.StatusPending, constant.StatusSuccess},
				StartAt:    &util.TimeNow,
				EndAt:      &util.TimeNow,
			},
		},
		{
			name: "ERROR: GetAggregateTransactions not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: GetAggregateTransactions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(errors.New("some error"))

			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)
			req := request
			if tc.request != nil {
				req = tc.request
			}
			transaction, err := repo.GetAggregateTransactions(ctx, req)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCalculatePendingBalance(t *testing.T) {
	paymentPendingBalance := float64(-1000.00)

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  float64
		wantErr   bool
	}{
		{
			name: "SUCCESS: CalculatePaymentPendingBalance",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*float64) = paymentPendingBalance
				})
			},
			expected: paymentPendingBalance,
			wantErr:  false,
		},
		{
			name: "ERROR: CalculatePaymentPendingBalance not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
				).Return(sql.ErrNoRows)
			},
			expected: 0,
			wantErr:  false,
		},
		{
			name: "ERROR: CalculatePaymentPendingBalance",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
				).Return(errors.New("some error"))
			},
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Setup the mock
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			balance, err := repo.CalculatePendingBalance(context.Background(), &orchestrator_model.GetAggregateRequest{
				MerchantID: util.ParseUUID("ca0b95b3-3b5d-4b46-a3e1-0034528ec49e"),
				AccountID:  util.ParseUUID("5e29a622-730c-4efd-b4b2-6360f5ccae1d"),
				StartAt:    util.ValueToPtr(time.Now().UTC().Add(-time.Hour)),
				EndAt:      util.ValueToPtr(time.Now().UTC()),
			})

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, balance)
			}
			mockMysql.AssertExpectations(t)
		})
	}
}

func TestGetAggregateTransactionByReference(t *testing.T) {
	aggregateResponse := &orchestrator_model.AggregateResponse{
		SumOfDebit:  0,
		SumOfCredit: 0,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		request   *orchestrator_model.GetSummaryTransactionByReferenceRequest
		expected  *orchestrator_model.AggregateResponse
		wantErr   bool
	}{
		{
			name: "SUCCESS: GetAggregateTransactions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					aggregateResponsePtr := args.Get(1).(*orchestrator_model.AggregateResponse)
					*aggregateResponsePtr = *aggregateResponse
				})
			},
			request: &orchestrator_model.GetSummaryTransactionByReferenceRequest{
				MerchantID: uuid.New(),
			},
			expected: aggregateResponse,
			wantErr:  false,
		},
		{
			name: "SUCCESS: GetAggregateTransactions with Status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					aggregateResponsePtr := args.Get(1).(*orchestrator_model.AggregateResponse)
					*aggregateResponsePtr = *aggregateResponse
				})
			},
			request: &orchestrator_model.GetSummaryTransactionByReferenceRequest{
				MerchantID:    uuid.New(),
				ReferenceID:   uuid.New().String(),
				ReferenceType: "PAYMENT",
				Status:        constant.StatusSuccess,
			},
			expected: aggregateResponse,
			wantErr:  false,
		},
		{
			name: "SUCCESS: GetAggregateTransactions with SettlementStatus",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					aggregateResponsePtr := args.Get(1).(*orchestrator_model.AggregateResponse)
					*aggregateResponsePtr = *aggregateResponse
				})
			},
			request: &orchestrator_model.GetSummaryTransactionByReferenceRequest{
				MerchantID:       uuid.New(),
				ReferenceID:      uuid.New().String(),
				ReferenceType:    "PAYMENT",
				SettlementStatus: constant.StatusSuccess,
			},
			expected: aggregateResponse,
			wantErr:  false,
		},
		{
			name: "SUCCESS: GetAggregateTransactions with Status and SettlementStatus",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					aggregateResponsePtr := args.Get(1).(*orchestrator_model.AggregateResponse)
					*aggregateResponsePtr = *aggregateResponse
				})
			},
			request: &orchestrator_model.GetSummaryTransactionByReferenceRequest{
				MerchantID:       uuid.New(),
				ReferenceID:      uuid.New().String(),
				ReferenceType:    "PAYMENT",
				Status:           constant.StatusSuccess,
				SettlementStatus: constant.StatusPending,
			},
			expected: aggregateResponse,
			wantErr:  false,
		},
		{
			name: "ERROR: GetAggregateTransactions not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			request: &orchestrator_model.GetSummaryTransactionByReferenceRequest{
				MerchantID: uuid.New(),
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: GetAggregateTransactions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.AggregateResponse"),
					constant.StringMockType(),
				).Return(errors.New("some error"))

			},
			request: &orchestrator_model.GetSummaryTransactionByReferenceRequest{
				MerchantID: uuid.New(),
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)
			transaction, err := repo.GetAggregateTransactionByReference(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetEarliestUpdatedAt(t *testing.T) {
	expectedTime := util.TimeNow
	request := &orchestrator_model.GetAggregateRequest{
		MerchantID: uuid.New(),
		AccountID:  uuid.New(),
		Statuses:   []string{constant.StatusPending, constant.StatusSuccess},
		StartAt:    &util.TimeNow,
		EndAt:      &util.TimeNow,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  time.Time
		wantErr   bool
		request   *orchestrator_model.GetAggregateRequest
	}{
		{
			name: "SUCCESS: GetEarliestUpdatedAt",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*time.Time"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					timePtr := args.Get(1).(*time.Time)
					*timePtr = expectedTime
				})
			},
			expected: expectedTime,
			wantErr:  false,
		},
		{
			name: "SUCCESS: GetEarliestUpdatedAt with AccountIDs",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*time.Time"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					timePtr := args.Get(1).(*time.Time)
					*timePtr = expectedTime
				})
			},
			expected: expectedTime,
			wantErr:  false,
			request: &orchestrator_model.GetAggregateRequest{
				MerchantID: uuid.New(),
				AccountIDs: []string{uuid.New().String()},
				Statuses:   []string{constant.StatusPending, constant.StatusSuccess},
				StartAt:    &util.TimeNow,
				EndAt:      &util.TimeNow,
			},
		},
		{
			name: "SUCCESS: GetEarliestUpdatedAt with pending settlement balance",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything, mock.MatchedBy(func(t *time.Time) bool { return t != nil && t.IsZero() }), constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*time.Time) = expectedTime
				})
			},
			expected: expectedTime,
			wantErr:  false,
			request: &orchestrator_model.GetAggregateRequest{
				MerchantID:               uuid.New(),
				AccountIDs:               []string{uuid.New().String()},
				StartAt:                  &util.TimeNow,
				EndAt:                    &util.TimeNow,
				PendingSettlementBalance: true,
			},
		},
		{
			name: "ERROR: GetEarliestUpdatedAt not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*time.Time"),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			expected: time.Time{},
			wantErr:  false,
		},
		{
			name: "ERROR: GetEarliestUpdatedAt",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*time.Time"),
					constant.StringMockType(),
				).Return(errors.New("some error"))
			},
			expected: time.Time{},
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)

			if tc.request == nil {
				tc.request = request
			}
			result, err := repo.GetEarliestUpdatedAt(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
			mockMysql.AssertExpectations(t)
		})
	}
}

func TestAccountTransactionRepositoryGetBulkAggregateTransactions(t *testing.T) {
	aggregateResponse := []*orchestrator_model.BulkAggregateResponse{
		{
			SumOfDebit:  0,
			SumOfCredit: 0,
		},
	}
	request := &orchestrator_model.BulkGetAggregateRequest{
		IncludeFeeIndirectDeduction: false,
		AccountClauses: []orchestrator_model.AccountsAggregationClause{
			{
				AccountID: uuid.NewString(),
				StartAt:   &util.TimeNow,
				EndAt:     &util.TimeNow,
			},
			{
				AccountID: uuid.NewString(),
				StartAt:   &util.TimeNow,
				EndAt:     &util.TimeNow,
			},
		},
		Statuses: []string{constant.StatusPending, constant.StatusSuccess},
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  []*orchestrator_model.BulkAggregateResponse
		wantErr   bool
		request   *orchestrator_model.BulkGetAggregateRequest
	}{
		{
			name: "SUCCESS: GetBulkAggregateTransactions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*orchestrator_model.BulkAggregateResponse"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					aggregateResponsePtr := args.Get(1).(*[]*orchestrator_model.BulkAggregateResponse)
					*aggregateResponsePtr = aggregateResponse
				})
			},
			expected: aggregateResponse,
			wantErr:  false,
		},
		{
			name: "SUCCESS: GetBulkAggregateTransactions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*orchestrator_model.BulkAggregateResponse"),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					aggregateResponsePtr := args.Get(1).(*[]*orchestrator_model.BulkAggregateResponse)
					*aggregateResponsePtr = aggregateResponse
				})
			},
			expected: aggregateResponse,
			wantErr:  false,
			request: &orchestrator_model.BulkGetAggregateRequest{
				AccountClauses: []orchestrator_model.AccountsAggregationClause{
					{
						AccountID: uuid.NewString(),
						StartAt:   &util.TimeNow,
						EndAt:     &util.TimeNow,
					},
				},
				Statuses: []string{constant.StatusPending, constant.StatusSuccess},
			},
		},
		{
			name: "ERROR: GetBulkAggregateTransactions not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*orchestrator_model.BulkAggregateResponse"),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: GetGetBulkAggregateTransactionsAggregateTransactions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*orchestrator_model.BulkAggregateResponse"),
					constant.StringMockType(),
				).Return(errors.New("some error"))

			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)
			req := request
			if tc.request != nil {
				req = tc.request
			}
			transaction, err := repo.GetBulkAggregateTransactions(ctx, req)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
