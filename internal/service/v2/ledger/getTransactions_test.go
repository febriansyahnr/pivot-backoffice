package ledgerService

import (
	"context"
	"errors"
	"testing"
	"time"

	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
)

func TestGetLedgerTransactions(t *testing.T) {

	testCases := []struct {
		Name      string
		Request   *ledger_model.GetLedgerTransactionRequest
		MockSetup func(accSvc *mockSvc.IAccountService,
			repo *mockRepo.IAccountTransactionRepository,
		)
		WantErr bool
	}{
		{
			Name: "SUCCESS: Get Ledger Transactions",
			Request: &ledger_model.GetLedgerTransactionRequest{
				AccountID: uuid.New(),
			},
			MockSetup: func(accSvc *mockSvc.IAccountService,
				repo *mockRepo.IAccountTransactionRepository,
			) {
				repo.On("GetLedgerRecords",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]*ledger_model.GetLedgerTransactionData{}, 100, nil)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Invalid datetime",
			Request: &ledger_model.GetLedgerTransactionRequest{
				AccountID: uuid.New(),
				StartDate: time.Now(),
				EndDate:   time.Now().AddDate(0, 0, -1),
			},
			MockSetup: func(accSvc *mockSvc.IAccountService,
				repo *mockRepo.IAccountTransactionRepository,
			) {
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Error get Ledger Transactions",
			Request: &ledger_model.GetLedgerTransactionRequest{
				AccountID: uuid.New(),
			},
			MockSetup: func(accSvc *mockSvc.IAccountService,
				repo *mockRepo.IAccountTransactionRepository,
			) {

				repo.On("GetLedgerRecords",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, 0, errors.New("error"))
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			accSvc := mockSvc.NewIAccountService(t)
			accTrxRepo := mockRepo.NewIAccountTransactionRepository(t)
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.MockSetup(accSvc, accTrxRepo)

			svc := New(loggerMock, accTrxRepo, nil, nil, nil, accSvc)
			_, err := svc.GetLedgerTransactions(context.Background(), tc.Request, &commonModel.Meta{
				PerPage: constant.DefaultPaginationPageSize,
			})
			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
