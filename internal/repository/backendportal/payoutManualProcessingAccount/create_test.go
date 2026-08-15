package payoutManualProcessingAccount_test

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payoutManualProcessingAccount"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/payoutManualProcessingAccount"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	data := &payoutManualProcessingAccountModel.PayoutManualProcessingAccount{
		UUID:          "uuid-123",
		MerchantID:    "merchant-123",
		BankCode:      "BCA",
		AccountNumber: "1234567890",
		Status:        constant.StatusActive,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Insert Data",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Insert to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					mock.AnythingOfType("*payoutManualProcessingAccount.PayoutManualProcessingAccount"),
				).Return(false, errors.New("insert error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockLogger, mockMysql)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payout_manual_processing_accounts")
			err := repo.Create(ctx, data)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
