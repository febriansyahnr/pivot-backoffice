package accounttransaction_repository

import (
	"context"
	"database/sql"
	"testing"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	reconciliation "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/reconciliation"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionByProcessorID(t *testing.T) {
	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr    bool
		wantResult *reconciliation.ReconTransactionModel
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*reconciliation.ReconTransactionModel)
					*result = reconciliation.ReconTransactionModel{
						UUID:    "test-uuid",
						Type:    "PAYMENT",
						Amount:  decimal.NewFromInt(1000),
						Channel: "test-channel",
					}
				}).Return(nil)
			},
			wantResult: &reconciliation.ReconTransactionModel{
				UUID:    "test-uuid",
				Type:    "PAYMENT",
				Amount:  decimal.NewFromInt(1000),
				Channel: "test-channel",
			},
		},
		{
			name: "ERROR: Data not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			wantResult: nil,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.ReconTransactionModel"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			result, err := repo.GetTransactionByProcessorID(context.Background(), "PAYMENT", "processor", "processor-id")

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantResult, result)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
