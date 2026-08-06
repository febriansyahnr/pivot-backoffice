package withdrawalService_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	log := mocks.NewILogger(t)
	mutex := redisMock.NewIMutexer(t)
	redis := redisMock.NewIRedisExt(t)
	rmq := rabbitmqExtMock.NewRabbitMQExt(t)
	disbursementSvc := serviceMocks.NewIDisbursementService(t)

	config := &config.WithdrawalConfig{
		MinAmount:           10_000,        // NOSONAR
		MaxAmount:           250_000_000,   // NOSONAR
		LimitOverbooking:    1_000_000_000, // NOSONAR
		LimitNonOverbooking: 300_000_000,   // NOSONAR
	}
	merchantConfig := &merchant.WithdrawalConfig{
		MinAmount: 50_000,      // NOSONAR
		MaxAmount: 500_000_000, // NOSONAR
	}

	userSvc := serviceMocks.NewIUserService(t)
	notificationSvc := serviceMocks.NewINotificationService(t)
	bankAccRepo := repoMocks.NewIBankAccountRepository(t)
	withdrawalRepo := repoMocks.NewIWithdrawalRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	snapCoreRepo := repoMocks.NewISnapCoreRepository(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := New(
		log, withdrawalRepo, orchestratorSvc, bankAccRepo, userSvc,
		WithRedisClient(redis),
		WithRabbitMQClient(rmq),
		WithWithdrawalConfig(config),
		WithSnapCoreRepository(snapCoreRepo),
		WithBankTransferConfig(disbursementSvc),
		WithMerchantRepository(merchantRepo),
		WithNotificationService(notificationSvc),
	)

	traceId := uuid.NewString()
	redsyncOptionMockType := mock.AnythingOfType("redsync.OptionFunc")

	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)
	ctxTx := context.WithValue(ctx, pdkConst.CtxTraceIdKey, &sql.Tx{})

	transactionId := uuid.New()
	availableBalance := 50_000.00
	withdrawalId := "f1c70284-d587-458b-9616-d2c3e8ade264"
	generalErrResult := pkgErrs.New(response.HttpErrInternal, fmt.Errorf(c.InternalErrorFmt, traceId))
	withdrawalRequest := &withdrawal.WithdrawalRequest{
		Type:        c.WithdrawalManual,
		AccountName: c.TypePayment,
		Amount:      50_000,
	}

	rmq.On(
		"PublishActivity", c.ValueCtxMockType(), mock.Anything, mock.Anything, c.StringMockType(), c.StringMockType(), mock.Anything,
	).Return(nil)
	log.On(
		"Info", mock.Anything, "Withdrawal transaction status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return()

	tests := []struct {
		name        string
		destination string
		request     *withdrawal.WithdrawalRequest
		setupMock   func()
		wantErr     error
		wantResult  *withdrawal.WithdrawalProcessResponse
	}{
		{
			name: "ERROR:Get transaction config",
			setupMock: func() {
				withdrawalRepo.On("GetTransactionConfig", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed while get withdrawal transaction config", mock.Anything).Once().Return()
			},
			wantErr: assert.AnError, // NOSONAR
		},
		{
			name: "ERROR:Get Merchant Balance",
			setupMock: func() {
				withdrawalRepo.On("GetTransactionConfig", mock.Anything, mock.Anything).Once().Return(nil, nil)
				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).Once().Return(float64(0), assert.AnError)
				log.On("Error", mock.Anything, "Get available merchant balance", mock.Anything).Once().Return()
			},
			request: &withdrawal.WithdrawalRequest{
				Type:         c.WithdrawalManual,
				AccountName:  c.TypePayment,
				Amount:       50_000,
				IsFullAmount: true,
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, fmt.Errorf(c.InternalErrorFmt, traceId)), // NOSONAR
		},
		{
			name:    "ERROR:The amount is less than the minimum transaction (default config)",
			request: &withdrawal.WithdrawalRequest{Amount: 9_500},
			setupMock: func() {
				withdrawalRepo.On("GetTransactionConfig", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf(TransactionAmountBelowMinLimitErrMessageFmt, util.ConvertFloatToCurrency(config.MinAmount))),
		},
		{
			name:    "ERROR:The amount is less than the minimum transaction (custom config)",
			request: &withdrawal.WithdrawalRequest{Amount: 9_500},
			setupMock: func() {
				withdrawalRepo.On("GetTransactionConfig", mock.Anything, mock.Anything).Once().Return(merchantConfig, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf(TransactionAmountBelowMinLimitErrMessageFmt, util.ConvertFloatToCurrency(merchantConfig.MinAmount))),
		},
		{
			name: "ERROR:Get bank account (automated withdrawal)",
			request: &withdrawal.WithdrawalRequest{
				Type:        c.WithdrawalAutomated,
				Amount:      50_000, // NOSONAR
				Destination: c.WithdrawalDestBankTransfer,
			},
			setupMock: func() {
				withdrawalRepo.On("GetTransactionConfig", mock.Anything, mock.Anything).Once().Return(nil, nil)
				bankAccRepo.On(
					"GetBankAccountValidation", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Get bank account validation", mock.Anything).Once().Return()
			},
			wantErr: generalErrResult,
		},
		{
			name: "ERROR:The amount is more than the maximum transaction (config merchant)",
			request: &withdrawal.WithdrawalRequest{
				Type:        c.WithdrawalManual,
				Destination: c.WithdrawalDestBankTransfer,
				Amount:      750_000_000,
			},
			setupMock: func() {
				withdrawalRepo.On("GetTransactionConfig", mock.Anything, mock.Anything).Once().Return(merchantConfig, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf(MerchantLimitExceededErrMessageFmt, util.ConvertFloatToCurrency(merchantConfig.MaxAmount))),
		},
		{
			name: "ERROR:The amount is more than the maximum transaction (overbooking)",
			request: &withdrawal.WithdrawalRequest{
				Type:        c.WithdrawalManual,
				Destination: c.WithdrawalDestBankTransfer,
				Amount:      1_100_000_000,
			},
			setupMock: func() {
				disbursementSvc.On(
					"IsBankcodeOverbookingChannelAllowed", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(true)
				withdrawalRepo.On(
					"GetTransactionConfig", mock.Anything, mock.Anything,
				).Once().Return(&merchant.WithdrawalConfig{
					MinAmount: 50_000,        // NOSONAR
					MaxAmount: 2_000_000_000, // NOSONAR
				}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf(BankLimitExceededErrMessageFmt, util.ConvertFloatToCurrency(config.LimitOverbooking))),
		},
		{
			name: "ERROR:The amount is more than the maximum transaction (non overbooking)",
			request: &withdrawal.WithdrawalRequest{
				Type:        c.WithdrawalManual,
				Destination: c.WithdrawalDestBankTransfer,
				Amount:      400_000_000,
			},
			setupMock: func() {
				disbursementSvc.On(
					"IsBankcodeOverbookingChannelAllowed", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false)
				withdrawalRepo.On("GetTransactionConfig", mock.Anything, mock.Anything).Return(merchantConfig, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf(BankLimitExceededErrMessageFmt, util.ConvertFloatToCurrency(config.LimitNonOverbooking))),
		},
		{
			name: "ERROR: Get bank accounts for open api request",
			request: &withdrawal.WithdrawalRequest{
				ReferenceID: "ref-id",
				Type:        c.WithdrawalManual,
				Destination: c.WithdrawalDestBankTransfer,
				Amount:      40_000_000,
				Source:      c.SourceOpenApi,
			},
			setupMock: func() {
				disbursementSvc.On(
					"IsBankcodeOverbookingChannelAllowed", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false)

				bankAccRepo.On("GetListBankAccount", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()
				log.On("Error", mock.Anything, "Get list bank account", mock.Anything).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "ERROR: Bank accounts not found",
			request: &withdrawal.WithdrawalRequest{
				ReferenceID: "ref-id",
				Type:        c.WithdrawalManual,
				Destination: c.WithdrawalDestBankTransfer,
				Amount:      40_000_000,
				Source:      c.SourceOpenApi,
			},
			setupMock: func() {
				disbursementSvc.On(
					"IsBankcodeOverbookingChannelAllowed", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false)

				bankAccRepo.On("GetListBankAccount", mock.Anything, mock.Anything).Return(nil, nil).Once()
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("bank account not found")),
		},
		{
			name: "ERROR: Get withdrawal by referenceId",
			request: &withdrawal.WithdrawalRequest{
				ReferenceID: "ref-id",
				Type:        c.WithdrawalManual,
				Destination: c.WithdrawalDestBalanceTransfer,
				Amount:      40_000_000,
			},
			setupMock: func() {
				disbursementSvc.On(
					"IsBankcodeOverbookingChannelAllowed", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false)

				withdrawalRepo.On("GetByReferenceId", mock.Anything, mock.Anything, mock.Anything).Return(nil, c.ErrSomeErrorForUnitTest).Once()
				log.On("Error", mock.Anything, "error when get withdrawal record by reference id", mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantErr: generalErrResult,
		},
		{
			name: "ERROR: ReferenceId already exist",
			request: &withdrawal.WithdrawalRequest{
				ReferenceID: "ref-id",
				Type:        c.WithdrawalManual,
				Destination: c.WithdrawalDestBalanceTransfer,
				Amount:      40_000_000,
			},
			setupMock: func() {
				disbursementSvc.On(
					"IsBankcodeOverbookingChannelAllowed", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false)

				withdrawalRepo.On("GetByReferenceId", mock.Anything, mock.Anything, mock.Anything).Return(&withdrawal.WithdrawalDetailResponse{}, nil).Once()
			},
			wantErr: pkgErrs.New(response.HttpErrDupCheck, fmt.Errorf("withdrawal with reference id %s already exists", "ref-id")),
		},
		{
			name: "ERROR:Get bank account (manual withdrawal)",
			setupMock: func() {
				disbursementSvc.On(
					"IsBankcodeOverbookingChannelAllowed", mock.Anything, mock.Anything, mock.Anything,
				).Return(true)
				bankAccRepo.On(
					"GetBankAccountValidation", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Get bank account validation", mock.Anything).Once().Return()
			},
			wantErr: generalErrResult,
		},
		{
			name: "ERROR:Bank account not found",
			setupMock: func() {
				bankAccRepo.On(
					"GetBankAccountValidation", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("bank account not found")), // NOSONAR
		},
		{
			name: "ERROR:Mutex lock",
			setupMock: func() {
				bankAccRepo.On(
					"GetBankAccountValidation", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&bankAccount.BankAccountResponse{
					BeneficiaryBankCode:    "002",
					BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
					BeneficiaryAccountNo:   "00000000001",
					BeneficiaryAccountName: "JOHN WICK",
				}, nil)
				redis.On("NewMutex", c.StringMockType(), redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType, redsyncOptionMockType).Return(mutex)

				mutex.On("LockContext", c.ValueCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Distributed lock for balance deduction", mock.Anything).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Begin transaction",
			setupMock: func() {
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)

				withdrawalRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Once().Return(false, c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Begin transaction", mock.Anything).Once().Return()
				log.On("Warn", mock.Anything, "Failed unlock distributed lock", mock.Anything).Once().Return()
			},
			wantErr: generalErrResult,
		},
		{
			name: "ERROR:Get available balance",
			setupMock: func() {
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				withdrawalRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxTx, nil)

				orchestratorSvc.On(
					"GetAvailableMerchantBalance", ctxTx, c.StringMockType(), c.StringMockType(),
				).Once().Return(0.00, c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Get available merchant balance", mock.Anything).Times(1).Return()

				withdrawalRepo.On("RollbackTransaction", ctxTx).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Rollback session transaction", mock.Anything).Times(1).Return()
			},
			wantErr: generalErrResult,
		},
		{
			name:    "ERROR:Insufficient balance",
			request: withdrawalRequest,
			setupMock: func() {
				withdrawalRepo.On("RollbackTransaction", ctxTx).Return(nil)

				orchestratorSvc.On("GetAvailableMerchantBalance", ctxTx, c.StringMockType(), c.StringMockType()).Once().Return(9_000.00, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrForbidden, c.ErrInsufficientBalance),
		},
		{
			name: "ERROR:Create withdrawal history",
			setupMock: func() {
				orchestratorSvc.On("GetAvailableMerchantBalance", ctxTx, c.StringMockType(), c.StringMockType()).Return(availableBalance, nil)

				withdrawalRepo.On("Create", ctxTx, mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Create withdrawal history", mock.Anything).Once().Return()
			},
			wantErr: generalErrResult,
		},
		{
			name:        "ERROR:Create ledger of withdrawal transaction",
			destination: c.WithdrawalDestBalanceTransfer,
			setupMock: func() {
				withdrawalRepo.On(
					"Create", ctxTx, mock.Anything,
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*withdrawal.Withdrawal)) = withdrawal.Withdrawal{Id: withdrawalId}
				}).Return(nil)

				orchestratorSvc.On("PostAccountTransaction", ctxTx, mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Create ledger of withdrawal transaction", mock.Anything).Once().Return()
			},
			wantErr: generalErrResult,
		},
		{
			name:        "ERROR:Create ledger of top up balance transaction",
			destination: c.WithdrawalDestBalanceTransfer,
			setupMock: func() {
				orchestratorSvc.On("PostAccountTransaction", ctxTx, mock.Anything).Times(1).Return(nil)
				orchestratorSvc.On("PostAccountTransaction", ctxTx, mock.Anything).Times(1).Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Create ledger of top up balance transaction", mock.Anything).Once().Return()
			},
			wantErr: generalErrResult,
		},
		{
			name:        "ERROR:Commit transaction",
			destination: c.WithdrawalDestBalanceTransfer,
			setupMock: func() {
				orchestratorSvc.On(
					"PostAccountTransaction", ctxTx, mock.Anything,
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*orchestrator_model.CreateAccountTransactionRequest)) = orchestrator_model.CreateAccountTransactionRequest{UUID: transactionId}
				}).Return(nil)

				withdrawalRepo.On("CommitTransaction", ctxTx).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Commit session transaction", mock.Anything).Once().Return()
			},
			wantErr: generalErrResult,
		},
		{
			name:        "SUCCESS:Withdrawal to balance transfer",
			destination: c.WithdrawalDestBalanceTransfer,
			setupMock: func() {
				withdrawalRepo.On("CommitTransaction", ctxTx).Return(nil)
			},
			wantResult: &withdrawal.WithdrawalProcessResponse{
				Id:          withdrawalId,
				Type:        c.WithdrawalManual,
				AccountName: c.TypePayment,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Status: c.StatusSuccess,
			},
		},
		{
			name: "ERROR:Snap bank transfer",
			setupMock: func() {
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(&snapCoreModel.BankTransferResponseData{UUID: uuid.NewString()}, c.ErrSomeErrorForUnitTest)

				withdrawalRepo.On(
					"UpdateMetadataById", c.ValueCtxMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Update withdrawal metadata (bank transfer)", mock.Anything).Once().Return()

				orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Update account transactions additional info", mock.Anything).Once().Return()

				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.PtrStringMockType(), c.PtrStringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Update status account transaction (failed)", mock.Anything).Once().Return()
			},
			wantResult: &withdrawal.WithdrawalProcessResponse{
				Id:          withdrawalId,
				Type:        c.WithdrawalManual,
				AccountName: c.TypePayment,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Status: c.StatusFailed,
				Reason: c.ReasonTypeOtherReason,
			},
		},
		{
			name: "ERROR:Update status withdrawal history and ledger",
			setupMock: func() {
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(&snapCoreModel.BankTransferResponseData{
					UUID:   uuid.NewString(),
					Status: c.SnapCoreBankTransferStatusSuccess,
				}, nil)
				withdrawalRepo.On("UpdateMetadataById", c.ValueCtxMockType(), c.StringMockType(), mock.Anything).Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType()).
					Return(nil)

				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.PtrStringMockType(), c.PtrStringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Update status account transaction (success)", mock.Anything).Once().Return()
			},
			wantResult: &withdrawal.WithdrawalProcessResponse{
				Id:          withdrawalId,
				Type:        c.WithdrawalManual,
				AccountName: c.TypePayment,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Status: c.StatusPending,
			},
		},
		{
			name: "ERROR:Update status withdrawal history and ledger & error get withdrawal by id",
			setupMock: func() {
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(&snapCoreModel.BankTransferResponseData{
					UUID:   uuid.NewString(),
					Status: c.SnapCoreBankTransferStatusSuccess,
				}, nil)
				withdrawalRepo.On("UpdateMetadataById", c.ValueCtxMockType(), c.StringMockType(), mock.Anything).Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType()).
					Return(nil)

				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.PtrStringMockType(), c.PtrStringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, "Update status account transaction (success)", mock.Anything).Once().Return()
			},
			wantResult: &withdrawal.WithdrawalProcessResponse{
				Id:          withdrawalId,
				Type:        c.WithdrawalManual,
				AccountName: c.TypePayment,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Status: c.StatusPending,
			},
		},
		{
			name: "ERROR: Invalid/Dormant Bank Account",
			request: &withdrawal.WithdrawalRequest{
				Type:        c.WithdrawalAutomated,
				AccountName: c.TypePayment,
				Amount:      50_000,
			},
			setupMock: func() {
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(&snapCoreModel.BankTransferResponseData{
					UUID:         uuid.NewString(),
					Status:       c.SnapCoreBankTransferStatusFailed,
					ResponseCode: "4031809",
				}, errors.New("error"))
				withdrawalRepo.On("UpdateMetadataById", c.ValueCtxMockType(), c.StringMockType(), mock.Anything).Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType()).
					Return(nil)

				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.PtrStringMockType(), c.PtrStringMockType(),
				).Once().Return(nil)

				notificationSvc.On(
					"SendFailedWithdrawalAlert", mock.Anything, mock.Anything,
				).Return(nil).Once()

			},
			wantResult: &withdrawal.WithdrawalProcessResponse{
				Id:          withdrawalId,
				Type:        c.WithdrawalAutomated,
				AccountName: c.TypePayment,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Status: c.StatusFailed,
				Reason: c.ReasonTypeBeneficiaryAccountReason,
			},
		},
		{
			name: "ERROR: Send Slack Alert for Invalid/Dormant Bank Account",
			request: &withdrawal.WithdrawalRequest{
				Type:        c.WithdrawalAutomated,
				AccountName: c.TypePayment,
				Amount:      50_000,
			},
			setupMock: func() {
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(&snapCoreModel.BankTransferResponseData{
					UUID:         uuid.NewString(),
					Status:       c.SnapCoreBankTransferStatusFailed,
					ResponseCode: "4031809",
				}, errors.New("error"))
				withdrawalRepo.On("UpdateMetadataById", c.ValueCtxMockType(), c.StringMockType(), mock.Anything).Return(nil)
				orchestratorSvc.On("UpdateProcessorAndReconReferenceByID", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType()).
					Return(nil)

				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.PtrStringMockType(), c.PtrStringMockType(),
				).Once().Return(nil)

				notificationSvc.On(
					"SendFailedWithdrawalAlert", mock.Anything, mock.Anything,
				).Return(errors.New("error")).Once()

				log.On("Error", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

			},
			wantResult: &withdrawal.WithdrawalProcessResponse{
				Id:          withdrawalId,
				Type:        c.WithdrawalAutomated,
				AccountName: c.TypePayment,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Status: c.StatusFailed,
				Reason: c.ReasonTypeBeneficiaryAccountReason,
			},
		},
		{
			name: "SUCCESS:Pending",
			setupMock: func() {
				snapCoreRepo.On(
					"BankTransfer", c.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(&snapCoreModel.BankTransferResponseData{
					UUID:   uuid.NewString(),
					Status: c.SnapCoreBankTransferStatusPending,
				}, nil)
			},
			wantResult: &withdrawal.WithdrawalProcessResponse{
				Id:          withdrawalId,
				Type:        c.WithdrawalManual,
				AccountName: c.TypePayment,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Status: c.StatusPending,
			},
		},
		{
			name: "SUCCESS:Success",
			setupMock: func() {
				snapCoreRepo.On(
					"BankTransfer",
					c.ValueCtxMockType(),
					mock.MatchedBy(func(req *snapCoreModel.BankTransferRequest) bool {
						return req.Remark == withdrawalId[24:]
					}),
					mock.Anything,
				).Return(&snapCoreModel.BankTransferResponseData{
					UUID:   uuid.NewString(),
					Status: c.SnapCoreBankTransferStatusSuccess,
				}, nil)
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.PtrStringMockType(), c.PtrStringMockType(),
				).Return(nil)
			},
			wantResult: &withdrawal.WithdrawalProcessResponse{
				Id:          withdrawalId,
				Type:        c.WithdrawalManual,
				AccountName: c.TypePayment,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Status: c.StatusSuccess,
			},
		},
		{
			name: "SUCCESS: Success Full Amount from Open API",
			request: &withdrawal.WithdrawalRequest{
				IsFullAmount:           true,
				Type:                   c.WithdrawalManual,
				AccountName:            c.TypePayment,
				DestinationAccountName: c.AccountNameDisbursement,
				ReferenceID:            "ref-id",
				MerchantId:             uuid.Max.String(),
				Amount:                 50_000,
				Source:                 c.SourceOpenApi,
			},
			setupMock: func() {
				orchestratorSvc.On("GetAvailableMerchantBalance", mock.Anything, mock.Anything, mock.Anything).Return(float64(50_000), nil).Once()

				bankAccRepo.On("GetListBankAccount", mock.Anything, mock.Anything).Return([]bankAccount.BankAccountResponse{
					{
						BeneficiaryBankCode:    "bank-code",
						BeneficiaryBankName:    "bank-name",
						BeneficiaryAccountNo:   "bank-account-number",
						BeneficiaryAccountName: "bank-account-holder-name",
					},
				}, nil).Once()

				withdrawalRepo.On("GetByReferenceId", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()

				snapCoreRepo.On(
					"BankTransfer",
					c.ValueCtxMockType(),
					mock.MatchedBy(func(req *snapCoreModel.BankTransferRequest) bool {
						return req.Remark == withdrawalId[24:]
					}),
					mock.Anything,
				).Return(&snapCoreModel.BankTransferResponseData{
					UUID:   uuid.NewString(),
					Status: c.SnapCoreBankTransferStatusSuccess,
				}, nil)
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.PtrStringMockType(), c.PtrStringMockType(),
				).Return(nil)

				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to send withdrawal final status", mock.Anything, mock.Anything).Once().Return()
			},
			wantResult: &withdrawal.WithdrawalProcessResponse{
				Id:          withdrawalId,
				Type:        c.WithdrawalManual,
				AccountName: c.TypePayment,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Status: c.StatusSuccess,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			if test.destination == "" {
				test.destination = c.WithdrawalDestBankTransfer
			}
			if test.request == nil {
				test.request = withdrawalRequest
			}
			test.request.Destination = test.destination

			result, err := service.Create(ctx, test.request)
			assert.Equal(t, test.wantErr, err)
			if test.wantResult != nil && result != nil {
				test.wantResult.CreatedAt, test.wantResult.UpdatedAt = result.CreatedAt, result.UpdatedAt
			}
			assert.Equal(t, test.wantResult, result)

			merchantRepo.AssertExpectations(t)
		})
	}
}
