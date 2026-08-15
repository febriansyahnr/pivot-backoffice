package accounttransaction_repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ledger"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestToDomainModel(t *testing.T) {
	timestamp := time.Now().UTC()
	testCases := []struct {
		Name     string
		Input    []*GetLedgerDBRecords
		Expected []*ledger_model.GetLedgerTransactionData
	}{
		{
			Name: "Success",
			Input: []*GetLedgerDBRecords{
				{
					ReferenceID:          "TRXVA1234567890",
					Credit:               12500,
					Debit:                0,
					Type:                 constant.TypePayment,
					Channel:              constant.ChannelVirtualAccount,
					Status:               constant.StatusSuccess,
					Remarks:              "Transaction VA - TRXVA1234567890",
					TransactionTimestamp: timestamp,
				},
			},
			Expected: []*ledger_model.GetLedgerTransactionData{
				{
					ReferenceID:          "TRXVA1234567890",
					Credit:               12500,
					Debit:                0,
					Type:                 constant.TypePayment,
					Channel:              constant.ChannelVirtualAccount,
					Status:               constant.StatusSuccess,
					Remarks:              "Transaction VA - TRXVA1234567890",
					TransactionTimestamp: timestamp,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result := toDomainModel(tc.Input)
			if len(result) != len(tc.Expected) {
				t.Errorf("Expected %v, got %v", tc.Expected, result)
			}

			for i := 0; i < len(result); i++ {
				assert.Equal(t, tc.Expected[i], result[i])
			}
		})
	}
}

func TestGetLedgerRecords(t *testing.T) {

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt, wg *sync.WaitGroup)
		input     *ledger_model.GetLedgerTransactionRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get with all filter ",
			input: &ledger_model.GetLedgerTransactionRequest{
				AccountID:     uuid.New(),
				Status:        constant.StatusSuccess,
				ReferenceType: constant.TypePayment,
				StartDate:     time.Now().UTC(),
				EndDate:       time.Now().UTC().AddDate(0, 0, 7),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, wg *sync.WaitGroup) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					wg.Done()
				})

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*accounttransaction_repository.GetLedgerDBRecords"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil).Run(func(args mock.Arguments) {
					wg.Done()
				})

			},
			wantErr: false,
		},
		{
			name: "ERROR: Error get & count data",
			input: &ledger_model.GetLedgerTransactionRequest{
				AccountID:     uuid.New(),
				Status:        constant.StatusSuccess,
				ReferenceType: constant.TypePayment,
				StartDate:     time.Now().UTC(),
				EndDate:       time.Now().UTC().AddDate(0, 0, 7),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, wg *sync.WaitGroup) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("error")).Run(func(args mock.Arguments) {
					wg.Done()
				})

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*accounttransaction_repository.GetLedgerDBRecords"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(errors.New("error")).Run(func(args mock.Arguments) {
					wg.Done()
				})

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(2)
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql, &wg)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			_, _, err := repo.GetLedgerRecords(ctx, tc.input, &commonModel.Meta{})

			wg.Wait()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
