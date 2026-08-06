package accounttransaction_repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountTransactionCreate(t *testing.T) {
	accountID := uuid.New()
	merchantID := uuid.New()
	accountTransactionToInsert := orchestrator_model.AccountTransaction{
		UUID:                 uuid.New(),
		ReferenceID:          "TRXVA1234567890",
		MerchantID:           merchantID,
		AccountID:            accountID,
		Currency:             "IDR",
		Credit:               12500,
		Debit:                0,
		Type:                 constant.ChannelVirtualAccount,
		Status:               constant.StatusSuccess,
		Remarks:              "Transaksi VA - 1234567890987654",
		TransactionTimestamp: time.Now(),
	}

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Insert Account Transaction to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*orchestrator_model.AccountTransaction"),
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
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*orchestrator_model.AccountTransaction"),
				).Return(false, fmt.Errorf("insert error"))

			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*orchestrator_model.AccountTransaction"),
				).Return(false, nil)

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
			err := repo.Create(ctx, &accountTransactionToInsert)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
