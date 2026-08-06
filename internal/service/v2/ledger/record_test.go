package ledgerService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
)

func TestRecordTransaction(t *testing.T) {
	initiatorAccountID := uuid.New()
	recipientAccountID := uuid.New()
	parentID := uuid.New()
	parentAccountID := uuid.New()

	testCases := []struct {
		Name      string
		Request   *ledger_model.CreateNewLedgerEntryRequest
		Ctx       context.Context
		MockSetup func(
			svc *serviceMocks.ILedgerMoneyFlowService,
			validatorSvc *serviceMocks.ILedgerValidatorService,
			repo *repoMocks.IAccountTransactionRepository,
		)
		WantErr bool
	}{
		{
			Name: "SUCCESS: Disbursement",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceDisbursement,
				TransferType:         constant.TransferTypePayOut,
				TransactionType:      constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "bank transfer disbursement",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientID:          uuid.New(),
				RecipientAccountID:   initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Disbursement from parent merchant",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceDisbursement,
				TransferType:         constant.TransferTypePayOut,
				TransactionType:      constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "bank transfer disbursement",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderID:             parentID,
				SenderAccountID:      parentAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Disbursement Top Up",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceDisbursement,
				TransferType:         constant.TransferTypePayIn,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "virtual account top up disbursement",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Disbursement Top Up to parent merchant",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceDisbursement,
				TransferType:         constant.TransferTypePayIn,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "virtual account top up disbursement",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      parentAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Disbursement Fee",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceDisbursement,
				TransferType:         constant.TransferTypePayOut,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "disbursement fee",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Payment",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferencePayment,
				TransferType:         constant.TransferTypePayIn,
				TransactionType:      constant.TypePayment,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "bank transfer payment",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Payment to parent merchant",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferencePayment,
				TransferType:         constant.TransferTypePayIn,
				TransactionType:      constant.TypePayment,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "bank transfer payment",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      parentAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet Transfer P2P",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceWallet,
				TransferType:         constant.TransferTypeP2P,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "wallet transfer p2p",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   recipientAccountID,
				SenderAccountID:      initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet TopUp",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceWallet,
				TransferType:         constant.TransferTypePayIn,
				TransactionType:      constant.TypeWalletTopUp,
				Channel:              "",
				Remarks:              "wallet top up",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientAccountID:   initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet withdrawal",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceWallet,
				TransferType:         constant.TransferTypePayOut,
				TransactionType:      constant.TypeWalletWithdrawal,
				Channel:              "",
				Remarks:              "wallet withdrawal",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				SenderAccountID:      initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: existing transaction in context",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceDisbursement,
				TransferType:         constant.TransferTypePayOut,
				TransactionType:      constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "existing tx",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientID:          uuid.New(),
				RecipientAccountID:   initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			Ctx: context.WithValue(context.Background(), pdkConstant.CtxSqlTx, util.ValueToPtr("existing-tx")),
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, _ *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Failed Validation",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:  uuid.NewString(),
				Usecase:      "unknown",
				TransferType: constant.TransferTypeP2P,
			},
			MockSetup: func(_ *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, _ *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(errors.New("error"))
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Create Transactions",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:  uuid.NewString(),
				Usecase:      constant.ReferenceDisbursement,
				TransferType: constant.TransferTypeP2P,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(errors.New("error"))
				repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Begin transaction fails",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceDisbursement,
				TransferType:         constant.TransferTypePayOut,
				TransactionType:      constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "begin tx error",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientID:          uuid.New(),
				RecipientAccountID:   initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(_ *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), errors.New("begin tx error"))
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Commit transaction fails",
			Request: &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:          uuid.NewString(),
				Usecase:              constant.ReferenceDisbursement,
				TransferType:         constant.TransferTypePayOut,
				TransactionType:      constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "commit tx error",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             constant.CurrencyIDR,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				RecipientID:          uuid.New(),
				RecipientAccountID:   initiatorAccountID,
				ParentID:             parentID,
				ParentAccountID:      parentAccountID,
			},
			MockSetup: func(svc *serviceMocks.ILedgerMoneyFlowService, validatorSvc *serviceMocks.ILedgerValidatorService, repo *repoMocks.IAccountTransactionRepository) {
				validatorSvc.On("ValidateTransaction", constant.ValueCtxMockType(), constant.StringMockType(), mock.AnythingOfType("*ledger_model.CreateNewLedgerEntryRequest")).Return(nil)
				repo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				svc.On("CreateTransactions", mock.Anything, mock.Anything).Return(nil)
				repo.On("CommitTransaction", mock.Anything).Return(errors.New("commit error"))
				repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			moneyFlowSvc := &serviceMocks.ILedgerMoneyFlowService{}
			validatorSvc := serviceMocks.NewILedgerValidatorService(t)
			repo := repoMocks.NewIAccountTransactionRepository(t)
			tc.MockSetup(moneyFlowSvc, validatorSvc, repo)

			ledgerSvc := New(mockLogger, repo, nil, nil, nil, nil)
			WithValidatorService(ledgerSvc, validatorSvc)
			WithMoneyFlowService(ledgerSvc, tc.Request.TransferType, moneyFlowSvc)

			ctx := tc.Ctx
			if ctx == nil {
				ctx = context.Background()
			}
			err := ledgerSvc.RecordTransaction(ctx, uuid.NewString(), tc.Request)

			if tc.WantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
