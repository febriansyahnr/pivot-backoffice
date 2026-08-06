package cancelMoneyFlowService

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTransactions(t *testing.T) {
	senderID := uuid.New()
	senderAccountID := uuid.New()
	receiverID := uuid.New()
	receiverAccountID := uuid.New()

	testCases := []struct {
		Name      string
		Request   ledger_model.CreateNewLedgerEntryRequest
		MockSetup func(
			accTrxRepo *mockRepo.IAccountTransactionRepository,
			ledgerSvc *mockSvc.ILedgerService,
		)
		WantErr bool
	}{
		{
			Name: "SUCCESS: Cancel",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeMerchantPayment,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCancel,
				SenderAccountID:      senderAccountID,
				RecipientAccountID:   receiverAccountID,
				SenderID:             senderID,
				RecipientID:          receiverID,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: true,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				accTrxRepo.On(
					"GetLedgerDetail", mock.Anything, mock.Anything,
				).Return([]orchestrator_model.AccountTransaction{
					{
						UUID:             uuid.New(),
						AccountID:        senderAccountID,
						ReferenceID:      "123",
						SettlementStatus: sql.NullString{String: constant.SettlementStatusPending, Valid: true},
					},
					{
						UUID:        uuid.New(),
						AccountID:   receiverAccountID,
						ReferenceID: "123",
						Debit:       10,
					}}, nil)

				accTrxRepo.On("BulkUpdateTransactions",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				accTrxRepo.On("Create",
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Get transactions",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeMerchantPayment,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCancel,
				SenderAccountID:      senderAccountID,
				RecipientAccountID:   receiverAccountID,
				SenderID:             senderID,
				RecipientID:          receiverID,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: true,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				accTrxRepo.On(
					"GetLedgerDetail", mock.Anything, mock.Anything,
				).Return(nil, errors.New("errors"))
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Empty transactions",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeMerchantPayment,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCancel,
				SenderAccountID:      senderAccountID,
				RecipientAccountID:   receiverAccountID,
				SenderID:             senderID,
				RecipientID:          receiverID,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: true,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				accTrxRepo.On(
					"GetLedgerDetail", mock.Anything, mock.Anything,
				).Return([]orchestrator_model.AccountTransaction{}, nil)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Not Allowed to cancel transaction",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeMerchantPayment,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCancel,
				SenderAccountID:      senderAccountID,
				RecipientAccountID:   receiverAccountID,
				SenderID:             senderID,
				RecipientID:          receiverID,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: true,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				accTrxRepo.On(
					"GetLedgerDetail", mock.Anything, mock.Anything,
				).Return([]orchestrator_model.AccountTransaction{
					{
						UUID:        uuid.New(),
						AccountID:   senderAccountID,
						ReferenceID: "123",
					},
					{
						UUID:        uuid.New(),
						AccountID:   receiverAccountID,
						ReferenceID: "123",
						Debit:       10,
					}}, nil)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Update Trx",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeMerchantPayment,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCancel,
				SenderAccountID:      senderAccountID,
				RecipientAccountID:   receiverAccountID,
				SenderID:             senderID,
				RecipientID:          receiverID,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: true,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				accTrxRepo.On(
					"GetLedgerDetail", mock.Anything, mock.Anything,
				).Return([]orchestrator_model.AccountTransaction{
					{
						UUID:             uuid.New(),
						AccountID:        senderAccountID,
						ReferenceID:      "123",
						SettlementStatus: sql.NullString{String: constant.SettlementStatusPending, Valid: true},
					},
					{
						UUID:        uuid.New(),
						AccountID:   receiverAccountID,
						ReferenceID: "123",
						Debit:       10,
					}}, nil)

				accTrxRepo.On("BulkUpdateTransactions",
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error"))
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Create Cancel Trx",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeMerchantPayment,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCancel,
				SenderAccountID:      senderAccountID,
				RecipientAccountID:   receiverAccountID,
				SenderID:             senderID,
				RecipientID:          receiverID,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: true,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				accTrxRepo.On(
					"GetLedgerDetail", mock.Anything, mock.Anything,
				).Return([]orchestrator_model.AccountTransaction{
					{
						UUID:             uuid.New(),
						AccountID:        senderAccountID,
						ReferenceID:      "123",
						SettlementStatus: sql.NullString{String: constant.SettlementStatusPending, Valid: true},
					},
					{
						UUID:        uuid.New(),
						AccountID:   receiverAccountID,
						ReferenceID: "123",
						Debit:       10,
					}}, nil)

				accTrxRepo.On("BulkUpdateTransactions",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				accTrxRepo.On("Create",
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error"))
			},
			WantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			accSvc := mockSvc.NewIAccountService(t)
			accTrxRepo := mockRepo.NewIAccountTransactionRepository(t)
			loggerMock, _ := logger.NewZapLogger(logger.Config{})
			ledgerSvc := mockSvc.NewILedgerService(t)
			merchantSvc := mockSvc.NewIMerchantService(t)

			tc.MockSetup(accTrxRepo, ledgerSvc)

			svc := New(loggerMock, accTrxRepo, accSvc, ledgerSvc, merchantSvc)
			ctx := context.WithValue(context.Background(), constant.CtxSetPendingTransaction, true)

			if err := svc.CreateTransactions(ctx, &tc.Request); tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

		})
	}

}
