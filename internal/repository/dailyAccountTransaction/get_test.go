package dailyAccountTransactionRepository_test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	dailyAccountTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/dailyAccountTransaction"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/dailyAccountTransaction"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestFindLatestByAccountIDAndTimezone(t *testing.T) {
	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr    bool
		wantResult *dailyAccountTransactionModel.DailyAccountTransaction
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), // ctx
					mock.AnythingOfType("*dailyAccountTransactionModel.DailyAccountTransaction"),
					constant.StringMockType(),
					"test-account-id",
					"Asia/Jakarta",
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*dailyAccountTransactionModel.DailyAccountTransaction)
					result.ID = "test-id"
					result.AccountID = "test-account-id"
					result.Timezone = "Asia/Jakarta"
					result.BegBalance = 1000.0
					result.CreditAmount = 500.0
					result.DebitAmount = 300.0
				}).Return(nil)
			},
			wantResult: &dailyAccountTransactionModel.DailyAccountTransaction{
				ID:           "test-id",
				AccountID:    "test-account-id",
				Timezone:     "Asia/Jakarta",
				BegBalance:   1000.0,
				CreditAmount: 500.0,
				DebitAmount:  300.0,
			},
		},
		{
			name: "ERROR: Mysql error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), // ctx
					mock.AnythingOfType("*dailyAccountTransactionModel.DailyAccountTransaction"),
					constant.StringMockType(),
					"test-account-id",
					"Asia/Jakarta",
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Data not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), // ctx
					mock.AnythingOfType("*dailyAccountTransactionModel.DailyAccountTransaction"),
					constant.StringMockType(),
					"test-account-id",
					"Asia/Jakarta",
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()
			accountID := "test-account-id"
			timezone := "Asia/Jakarta"

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			result, err := repo.FindLatestByAccountIDAndTimezone(ctx, accountID, timezone)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				fmt.Println(result, tc.wantResult)
				assert.Equal(t, tc.wantResult, result)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
