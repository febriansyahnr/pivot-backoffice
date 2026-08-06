package ledgerService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTransactions(t *testing.T) {
	request := &ledger_model.UpdateLedgerEntryRequest{
		ReferenceID: uuid.New(),
		Usecase:     constant.ReferenceDisbursement,
	}

	testCases := []struct {
		Name      string
		Request   *ledger_model.UpdateLedgerEntryRequest
		MockSetup func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository)
		WantErr   bool
	}{
		{
			Name: "SUCCESS: Disbursement success Update",
			Request: &ledger_model.UpdateLedgerEntryRequest{
				ReferenceID:       uuid.New(),
				Usecase:           constant.ReferenceDisbursement,
				Status:            constant.StatusSuccess,
				ReasonDescription: "",
				ReasonType:        "",
			},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Disbursement fail Update",
			Request: &ledger_model.UpdateLedgerEntryRequest{
				ReferenceID:       uuid.New(),
				Usecase:           constant.ReferenceDisbursement,
				Status:            constant.StatusFailed,
				ReasonDescription: "Transactions failed",
				ReasonType:        "Unknown",
			},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet withdrawal success update",
			Request: &ledger_model.UpdateLedgerEntryRequest{
				ReferenceID:       uuid.New(),
				Usecase:           constant.ReferenceWallet,
				Status:            constant.StatusSuccess,
				ReasonDescription: "",
				ReasonType:        "",
			},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet withdrawal fail update",
			Request: &ledger_model.UpdateLedgerEntryRequest{
				ReferenceID:       uuid.New(),
				Usecase:           constant.ReferenceWallet,
				Status:            constant.StatusFailed,
				ReasonDescription: "wallet withdrawal failed",
				ReasonType:        "UNKNOWN",
			},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name:    "ERROR: Failed Validation",
			Request: &ledger_model.UpdateLedgerEntryRequest{},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
			},
			WantErr: true,
		},
		{
			Name:    "ERROR: Update transactions",
			Request: request,
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {

				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accTrxRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			accSvc := serviceMocks.NewIAccountService(t)
			tc.MockSetup(accSvc, accTrxRepo)

			svc := New(nil, accTrxRepo, nil, nil, nil, accSvc)
			err := svc.UpdateTransaction(context.Background(), tc.Request)
			if tc.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBulkUpdateLedgerEntry(t *testing.T) {
	request := &ledger_model.BulkUpdateLedgerEntryRequest{
		ReferenceID: uuid.New(),
		Requests: []*ledger_model.UpdateLedgerEntryRequest{
			{
				Usecase: constant.ReferenceDisbursement,
			},
		},
	}

	testCases := []struct {
		Name      string
		Request   *ledger_model.BulkUpdateLedgerEntryRequest
		MockSetup func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository)
		WantErr   bool
	}{
		{
			Name: "SUCCESS: Disbursement success Update",
			Request: &ledger_model.BulkUpdateLedgerEntryRequest{
				ReferenceID: uuid.New(),
				Requests: []*ledger_model.UpdateLedgerEntryRequest{
					{
						Usecase:           constant.ReferenceDisbursement,
						Status:            constant.StatusSuccess,
						ReasonDescription: "",
						ReasonType:        "",
					},
				},
			},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Disbursement fail Update",
			Request: &ledger_model.BulkUpdateLedgerEntryRequest{
				ReferenceID: uuid.New(),
				Requests: []*ledger_model.UpdateLedgerEntryRequest{
					{
						Usecase:           constant.ReferenceDisbursement,
						Status:            constant.StatusFailed,
						ReasonDescription: "Transactions failed",
						ReasonType:        "Unknown",
					},
				},
			},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet withdrawal success update",
			Request: &ledger_model.BulkUpdateLedgerEntryRequest{
				ReferenceID: uuid.New(),
				Requests: []*ledger_model.UpdateLedgerEntryRequest{
					{
						Usecase:           constant.ReferenceWallet,
						Status:            constant.StatusSuccess,
						ReasonDescription: "",
						ReasonType:        "",
					},
				},
			},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet withdrawal fail update",
			Request: &ledger_model.BulkUpdateLedgerEntryRequest{
				ReferenceID: uuid.New(),
				Requests: []*ledger_model.UpdateLedgerEntryRequest{
					{
						Usecase:           constant.ReferenceWallet,
						Status:            constant.StatusFailed,
						ReasonDescription: "wallet withdrawal failed",
						ReasonType:        "UNKNOWN",
					},
				},
			},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name:    "ERROR: Failed Validation",
			Request: &ledger_model.BulkUpdateLedgerEntryRequest{},
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
			},
			WantErr: true,
		},
		{
			Name:    "ERROR: Begin Transaction",
			Request: request,
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), errors.New("error"))
			},
			WantErr: true,
		},
		{
			Name:    "ERROR: Update transactions",
			Request: request,
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(errors.New("error"))
				repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			WantErr: true,
		},
		{
			Name:    "ERROR: Rollback transactions",
			Request: request,
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(errors.New("error"))
				repo.On("RollbackTransaction", mock.Anything).Return(errors.New("error"))
			},
			WantErr: true,
		},
		{
			Name:    "ERROR: Commit transactions",
			Request: request,
			MockSetup: func(accSvc *serviceMocks.IAccountService, repo *repositoryMocks.IAccountTransactionRepository) {
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				repo.On("UpdateTransactionsStatus", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(errors.New("error"))
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accTrxRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			accSvc := serviceMocks.NewIAccountService(t)
			logger := logger.NewSlogger(logger.Config{})
			tc.MockSetup(accSvc, accTrxRepo)

			svc := New(logger, accTrxRepo, nil, nil, nil, accSvc)
			err := svc.BulkUpdateLedgerEntry(context.Background(), tc.Request)
			if tc.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
