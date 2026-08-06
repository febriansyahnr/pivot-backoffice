package p2pMoneyFlowService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rabbitmqExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTransactions(t *testing.T) {
	senderID := uuid.New()
	senderAccountID := uuid.New()
	parentID := uuid.New()
	parentAccountID := uuid.New()
	recipientID := uuid.New()
	recipientAccountID := uuid.New()

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
			Name: "SUCCESS: Create Wallet P2P Requests via Parent",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
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
			Name: "SUCCESS: Create Wallet P2P Requests to Parent Merchant",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				RecipientAccountID:   parentAccountID,
				RecipientID:          parentID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					mock.Anything,
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
			Name: "SUCCESS: Create Wallet P2P Requests From Parent Merchant",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      parentAccountID,
				SenderID:             parentID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					mock.Anything,
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
			Name: "ERROR: No recipient ID",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
				RecipientAccountID:   uuid.Nil,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
			},
			WantErr: true,
		},
		{
			Name: "ERROR: get Balance",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					senderAccountID,
				).Return(nil, errors.New("errors"))
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Insufficient Balance",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					senderAccountID,
				).Return(&ledger_model.LedgerBalance{Balance: 100, Currency: constant.CurrencyIDR}, nil)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Insufficient Balance + fee amount",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               99,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
				Fee: ledger_model.FeeRequest{
					Amount:             2,
					RecipientAccountID: parentAccountID,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					senderAccountID,
				).Return(&ledger_model.LedgerBalance{Balance: 100, Currency: constant.CurrencyIDR}, nil)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Invalid Request",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet + "invalid",
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      parentAccountID,
				ParentID:             parentID,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					mock.Anything,
				).Return(&ledger_model.LedgerBalance{Balance: 1000, Currency: constant.CurrencyIDR}, nil)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Bulk insert",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.WalletTrxTopUpType,
				Channel:              constant.ChannelManualTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ParentAccountID:      senderAccountID,
				RecipientAccountID:   recipientAccountID,
				RecipientID:          recipientID,
				Fee: ledger_model.FeeRequest{
					Amount:             1000,
					Channel:            constant.ChannelManualTransfer,
					RecipientID:        recipientID,
					RecipientAccountID: recipientAccountID,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				ledgerSvc.On("GetLedgerBalance",
					mock.Anything,
					mock.Anything,
				).Return(&ledger_model.LedgerBalance{Balance: 40000, Currency: constant.CurrencyIDR}, nil)

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

			accSvc := mockSvc.NewIAccountService(t)
			accTrxRepo := mockRepo.NewIAccountTransactionRepository(t)
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
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

func TestMarkForMerchantSettlement(t *testing.T) {
	log := loggerMock.NewILogger(t)
	rmq := rabbitmqExtMocks.NewRabbitMQExt(t)

	service := &P2PMoneyFlowService{
		logger: log, queues: rmq,
	}
	assert.NotNil(t, New(log, nil, nil, nil, nil, WithRabbitMQClient(rmq)))

	walletUserId := "01a4af14-4f3e-4b1f-9a25-7631d9e997d1"
	merchantId := "f22a47e1-38b0-4ac4-8f2a-3a6aeff0eecb"

	transactions := []*orchestratorModel.AccountTransaction{
		{AccountID: util.ParseUUID(walletUserId)},
		{AccountID: util.ParseUUID(merchantId), Type: constant.WalletTrxMerchantPaymentType},
		{AccountID: util.ParseUUID(merchantId), Type: constant.TypeFee},
	}

	tests := []struct {
		name         string
		request      *ledger_model.CreateNewLedgerEntryRequest
		transactions []*orchestratorModel.AccountTransaction
		setupMock    func()
	}{
		{
			name: "Empty transaction list",
		},
		{
			name: "Other transaction types",
			request: &ledger_model.CreateNewLedgerEntryRequest{
				TransactionType: "Others",
			},
			transactions: transactions,
		},
		{
			name: "Success",
			request: &ledger_model.CreateNewLedgerEntryRequest{
				RecipientAccountID: util.ParseUUID(merchantId),
				TransactionType:    constant.WalletTrxMerchantPaymentType,
			},
			transactions: transactions,
			setupMock: func() {
				rmq.On("PublishForSettlementProcess", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			fn := service.markForMerchantSettlement(test.request, test.transactions)
			assert.NotNil(t, fn)

			fn(context.Background())

			log.AssertExpectations(t)
			rmq.AssertExpectations(t)
		})
	}
}
