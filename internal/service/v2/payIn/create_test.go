package payInMoneyFlowService

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
	parentAccountID := uuid.New()
	parentID := uuid.New()
	recipientAccountID := uuid.New()
	recipientID := uuid.New()

	testCases := []struct {
		Name      string
		Request   ledger_model.CreateNewLedgerEntryRequest
		MockSetup func(
			accTrxRepo *mockRepo.IAccountTransactionRepository,
		)
		WantErr bool
	}{
		{
			Name: "SUCCESS: Create PayIn Requests via parent account",
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
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository) {
				accTrxRepo.On("BulkInsert",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create PayIn Requests directly",
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
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
				RecipientAccountID:   parentAccountID,
				RecipientID:          parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository) {

				accTrxRepo.On("BulkInsert",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Payment PayIn Request",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePayment,
				TransactionType:      constant.TypePayment,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository) {

				accTrxRepo.On("BulkInsert",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Wallet Withdrawal PayIn Request",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository) {
				accTrxRepo.On("BulkInsert",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Incorrect Requests",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      "Invalid Top Up",
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          uuid.Nil,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository) {
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Bulk Insert",
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
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository) {
				accTrxRepo.On("BulkInsert",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error"))
			},
			WantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			merchSvc := mockSvc.NewIMerchantService(t)
			accSvc := mockSvc.NewIAccountService(t)
			accTrxRepo := mockRepo.NewIAccountTransactionRepository(t)
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.MockSetup(accTrxRepo)

			svc := New(loggerMock, accTrxRepo, accSvc, merchSvc)
			err := svc.CreateTransactions(context.Background(), &tc.Request)
			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

		})
	}
}
