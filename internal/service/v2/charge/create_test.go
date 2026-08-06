package chargeMoneyFlowService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rabbitmqExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
			Name: "SUCCESS: Create Charge Requests",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: true,
				},
			},
			MockSetup: func(accTrxRepo *mockRepo.IAccountTransactionRepository, ledgerSvc *mockSvc.ILedgerService) {
				accTrxRepo.On("BulkInsert",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Charge Requests bypass balance checks",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: false,
				},
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
			Name: "ERROR: Get Balance",
			Request: ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderAccountID:      senderAccountID,
				SenderID:             senderID,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: false,
				},
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
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: false,
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

	// Test Private Function
	service := &ChargeMoneyFlowService{
		logger: log, queues: rmq,
	}
	assert.NotNil(t, New(log, nil, nil, nil, nil, WithRabbitMQClient(rmq)))

	transactions := []*orchestratorModel.AccountTransaction{{}}

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
			name: "Top up fee",
			request: &ledger_model.CreateNewLedgerEntryRequest{
				TransactionType: constant.TypeFee, SenderAdditionalInfo: map[string]any{},
			},
			transactions: transactions,
		},
		{
			name: "Bill payment fee",
			request: &ledger_model.CreateNewLedgerEntryRequest{
				TransactionType: constant.TypeFee, SenderAdditionalInfo: struct{}{},
			},
			transactions: transactions,
		},
		{
			name: "Some error",
			request: &ledger_model.CreateNewLedgerEntryRequest{
				TransactionType: constant.TypeFee, SenderAdditionalInfo: map[string]any{"referenceType": constant.WalletTrxMerchantPaymentType},
			},
			transactions: transactions,
			setupMock: func() {
				rmq.On(
					"PublishForSettlementProcess", mock.Anything, mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
				log.On(
					"Error", mock.Anything, "Failed while publish pending settlement process (charge)", mock.Anything,
				).Once().Return()
			},
		},
		{
			name: "Success",
			request: &ledger_model.CreateNewLedgerEntryRequest{
				TransactionType: constant.TypeFee, SenderAdditionalInfo: map[string]any{"referenceType": constant.WalletTrxMerchantPaymentType},
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
