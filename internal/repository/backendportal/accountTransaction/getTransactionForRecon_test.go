package accounttransaction_repository

import (
	"context"
	"database/sql"
	"strings"
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

func TestGetTransactionForRecon(t *testing.T) {
	amount := decimal.NewFromInt(10000)
	testCases := []struct {
		name      string
		input     *reconciliation.ReconTransactionQuery
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get transaction data",
			input: &reconciliation.ReconTransactionQuery{
				ReferenceID:     "ref123",
				Amount:          amount,
				TransactionDate: time.Now(),
				Reference:       strings.ToUpper(constant.TagPayment),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("time.Time"),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*reconciliation.ReconTransactionModel)
					*result = reconciliation.ReconTransactionModel{
						Type:   "PAYMENT",
						Amount: amount,
					}
				}).Return(nil)
			},
		},
		{
			name: "SUCCESS: Get transaction data with time duration",
			input: &reconciliation.ReconTransactionQuery{
				ReferenceID:      "ref123",
				Amount:           amount,
				TransactionDate:  time.Now(),
				Reference:        strings.ToUpper(constant.TagPayment),
				WithTimeDuration: true,
				Duration:         1 * time.Minute,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*reconciliation.ReconTransactionModel)
					*result = reconciliation.ReconTransactionModel{
						Type:   "PAYMENT",
						Amount: amount,
					}
				}).Return(nil)
			},
		},
		{
			name: "ERROR: Database error",
			input: &reconciliation.ReconTransactionQuery{
				ReferenceID:     "ref123",
				Amount:          amount,
				TransactionDate: time.Now(),
				Reference:       strings.ToUpper(constant.TagPayment),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: No data found",
			input: &reconciliation.ReconTransactionQuery{
				ReferenceID:     "ref123",
				Amount:          amount,
				TransactionDate: time.Now(),
				Reference:       strings.ToUpper(constant.TagPayment),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS: Get transaction with QR channel",
			input: &reconciliation.ReconTransactionQuery{
				ReferenceID:     "ref123",
				Amount:          amount,
				TransactionDate: time.Now(),
				Reference:       strings.ToUpper(constant.TagPayment),
				Channel:         "QR",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					constant.StringMockType(), // QRIS
					constant.StringMockType(), // QR
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*reconciliation.ReconTransactionModel)
					*result = reconciliation.ReconTransactionModel{
						Type:    "PAYMENT",
						Amount:  amount,
						Channel: "QR",
					}
				}).Return(nil)
			},
		},
		{
			name: "SUCCESS: Get transaction with QRIS channel",
			input: &reconciliation.ReconTransactionQuery{
				ReferenceID:     "ref123",
				Amount:          amount,
				TransactionDate: time.Now(),
				Reference:       strings.ToUpper(constant.TagPayment),
				Channel:         "QRIS",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					constant.StringMockType(), // QRIS
					constant.StringMockType(), // QR
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*reconciliation.ReconTransactionModel)
					*result = reconciliation.ReconTransactionModel{
						Type:    "PAYMENT",
						Amount:  amount,
						Channel: "QRIS",
					}
				}).Return(nil)
			},
		},
		{
			name: "SUCCESS: Get transaction with Facilitator settlement model",
			input: &reconciliation.ReconTransactionQuery{
				ReferenceID:     "ref123",
				Amount:          amount,
				TransactionDate: time.Now(),
				Reference:       strings.ToUpper(constant.TagPayment),
				SettlementModel: constant.PaymentMethodChannelTypeFacilitator,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*reconciliation.ReconTransactionModel)
					*result = reconciliation.ReconTransactionModel{
						Type:   "PAYMENT",
						Amount: amount,
					}
				}).Return(nil)
			},
		},
		{
			name: "SUCCESS: Get transaction with QRIS and Facilitator",
			input: &reconciliation.ReconTransactionQuery{
				ReferenceID:     "ref123",
				Amount:          amount,
				TransactionDate: time.Now(),
				Reference:       strings.ToUpper(constant.TagPayment),
				Channel:         "QRIS",
				SettlementModel: constant.PaymentMethodChannelTypeFacilitator,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					constant.StringMockType(), // QRIS
					constant.StringMockType(), // QR
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*reconciliation.ReconTransactionModel)
					*result = reconciliation.ReconTransactionModel{
						Type:    "PAYMENT",
						Amount:  amount,
						Channel: "QRIS",
					}
				}).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			_, err := repo.GetTransactionForRecon(context.Background(), tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
