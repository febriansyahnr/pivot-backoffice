package accounttransaction_repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/reconciliation"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTotalPaymentAmount(t *testing.T) {
	now := time.Now()
	amount1 := decimal.NewFromInt(10000)
	amount2 := decimal.NewFromInt(20000)

	testCases := []struct {
		name           string
		input          *reconciliation.PaymentTotalAmountQuery
		mockSetup      func(mysqlMock *mysqlMocks.IMySqlExt)
		expectedResult *reconciliation.PaymentTotalAmountResult
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "SUCCESS: Get total payment amount for virtual account",
			input: &reconciliation.PaymentTotalAmountQuery{
				ReferenceIDs: []string{"ref123", "ref456"},
				Channel:      constant.ChannelVirtualAccount,
				StartTime:    now.Add(-24 * time.Hour),
				EndTime:      now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				transactions := []reconciliation.ReconTransactionModel{
					{
						ProcessorReferenceNumber: "ref123",
						Amount:                   amount1,
					},
					{
						ProcessorReferenceNumber: "ref456",
						Amount:                   amount2,
					},
				}
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]reconciliation.ReconTransactionModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*[]reconciliation.ReconTransactionModel)
					*result = transactions
				}).Return(nil)
			},
			expectedResult: &reconciliation.PaymentTotalAmountResult{
				"ref123": amount1,
				"ref456": amount2,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Empty result when no transactions found",
			input: &reconciliation.PaymentTotalAmountQuery{
				ReferenceIDs: []string{"ref123"},
				Channel:      constant.ChannelVirtualAccount,
				StartTime:    now.Add(-24 * time.Hour),
				EndTime:      now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]reconciliation.ReconTransactionModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*[]reconciliation.ReconTransactionModel)
					*result = []reconciliation.ReconTransactionModel{}
				}).Return(nil)
			},
			expectedResult: &reconciliation.PaymentTotalAmountResult{},
			wantErr:        false,
		},
		{
			name: "SUCCESS: Multiple payments with same reference get aggregated",
			input: &reconciliation.PaymentTotalAmountQuery{
				ReferenceIDs: []string{"ref123"},
				Channel:      constant.ChannelVirtualAccount,
				StartTime:    now.Add(-24 * time.Hour),
				EndTime:      now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				transactions := []reconciliation.ReconTransactionModel{
					{
						ProcessorReferenceNumber: "ref123",
						Amount:                   amount1,
					},
					{
						ProcessorReferenceNumber: "ref123",
						Amount:                   amount2,
					},
				}
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]reconciliation.ReconTransactionModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*[]reconciliation.ReconTransactionModel)
					*result = transactions
				}).Return(nil)
			},
			expectedResult: &reconciliation.PaymentTotalAmountResult{
				"ref123": amount1.Add(amount2),
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Duration > 3 days should limit to 3 days",
			input: &reconciliation.PaymentTotalAmountQuery{
				ReferenceIDs: []string{"ref123"},
				Channel:      constant.ChannelVirtualAccount,
				StartTime:    now.Add(-7 * 24 * time.Hour), // 7 days ago
				EndTime:      now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				transactions := []reconciliation.ReconTransactionModel{
					{
						ProcessorReferenceNumber: "ref123",
						Amount:                   amount1,
					},
				}
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]reconciliation.ReconTransactionModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*[]reconciliation.ReconTransactionModel)
					*result = transactions
				}).Return(nil)
			},
			expectedResult: &reconciliation.PaymentTotalAmountResult{
				"ref123": amount1,
			},
			wantErr: false,
		},
		{
			name: "ERROR: Unsupported channel",
			input: &reconciliation.PaymentTotalAmountQuery{
				ReferenceIDs: []string{"ref123"},
				Channel:      constant.ChannelQris,
				StartTime:    now.Add(-24 * time.Hour),
				EndTime:      now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// No mock setup needed as it should fail before database call
			},
			expectedResult: &reconciliation.PaymentTotalAmountResult{},
			wantErr:        true,
			expectedErrMsg: fmt.Sprintf("channel %s is not supported", constant.ChannelQris),
		},
		{
			name: "ERROR: Database error",
			input: &reconciliation.PaymentTotalAmountQuery{
				ReferenceIDs: []string{"ref123"},
				Channel:      constant.ChannelVirtualAccount,
				StartTime:    now.Add(-24 * time.Hour),
				EndTime:      now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]reconciliation.ReconTransactionModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: No data found returns nil",
			input: &reconciliation.PaymentTotalAmountQuery{
				ReferenceIDs: []string{"ref123"},
				Channel:      constant.ChannelVirtualAccount,
				StartTime:    now.Add(-24 * time.Hour),
				EndTime:      now,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]reconciliation.ReconTransactionModel"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			expectedResult: nil,
			wantErr:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			result, err := repo.GetTotalPaymentAmount(context.Background(), tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				if tc.expectedResult == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, *tc.expectedResult, *result)
				}
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
