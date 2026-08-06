package payoutMoneyFlowService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
)

func TestCreateTransactions(t *testing.T) {
	senderID := uuid.New()
	senderAccountID := uuid.New()
	parentID := uuid.New()
	parentAccountID := uuid.New()

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
			Name: "SUCCESS: Create Payout Requests via parent account",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					senderAccountID,
				).Return(&ledger_model.LedgerBalance{Balance: 1000, Currency: constant.CurrencyIDR}, nil)

				accTrxRepo.On("BulkInsert",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Disbursement Payout Requests",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      senderAccountID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					senderAccountID,
				).Return(&ledger_model.LedgerBalance{Balance: 1000, Currency: constant.CurrencyIDR}, nil)

				accTrxRepo.On("BulkInsert",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Wallet Payout Requests",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletWithdrawal,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					senderAccountID,
				).Return(&ledger_model.LedgerBalance{Balance: 1000, Currency: constant.CurrencyIDR}, nil)

				accTrxRepo.On("BulkInsert",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: get Balance",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {

				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					senderAccountID,
				).Return(&ledger_model.LedgerBalance{Balance: 1000, Currency: constant.CurrencyIDR}, errors.New("error"))
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Insufficient Balance",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {

				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					mock.Anything,
				).Return(&ledger_model.LedgerBalance{Balance: 100, Currency: constant.CurrencyIDR}, nil)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Insufficient Balance + Fee Amount",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               100,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
				Fee: ledger_model.FeeRequest{
					Amount:             2,
					RecipientAccountID: parentAccountID,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {

				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					mock.Anything,
				).Return(&ledger_model.LedgerBalance{Balance: 100, Currency: constant.CurrencyIDR}, nil)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Invalid Request",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             uuid.Nil,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Bulk insert",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					senderAccountID,
				).Return(&ledger_model.LedgerBalance{Balance: 1000, Currency: constant.CurrencyIDR}, nil)

				accTrxRepo.On("BulkInsert",
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
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			ledgerSvc := mockSvc.NewILedgerService(t)
			merchantSvc := mockSvc.NewIMerchantService(t)

			tc.MockSetup(accTrxRepo, ledgerSvc)

			svc := New(loggerMock, accTrxRepo, accSvc, ledgerSvc, merchantSvc)
			err := svc.CreateTransactions(context.Background(), &tc.Request)
			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

		})
	}
}
