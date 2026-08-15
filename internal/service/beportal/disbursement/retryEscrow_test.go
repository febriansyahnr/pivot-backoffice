package disbursementService

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	routingProcessorModelEscrowBalance "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/escrowBalance"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryDueToInsufficientEscrowFund(t *testing.T) {
	type mocker struct {
		disbursementRepo               *repositoryMocks.IDisbursementRepository
		snapCoreRepo                   *repositoryMocks.ISnapCoreRepository
		bankAccountRepo                *repositoryMocks.IBankAccountRepository
		accountTransactionRepo         *repositoryMocks.IAccountTransactionRepository
		orchestratorSvc                *serviceMocks.IOrchestratorService
		beneficiaryAccSvc              *serviceMocks.IBeneficiaryAccountService
		rmqExt                         *rabbitMqMocks.RabbitMQExt
		forbiddenUsecaseSvc            *serviceMocks.IMerchantForbiddenUseCaseService
		feeSvc                         *serviceMocks.IFeeService
		routingProcessorSvc            *serviceMocks.IRoutingProcessorService
		statusHistoriesRepo            *repositoryMocks.IStatusHistoriesRepository
		payoutManualProcessingAcctRepo *repositoryMocks.IPayoutManualProcessingAccountRepository
		rdb                            *redisMock.IRedisExt
	}

	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}
	rdb := redisMock.NewIRedisExt(t)
	mutexLock := redisMock.NewIMutexer(t)
	disbursementID := uuid.NewString()
	merchantID := uuid.NewString()
	transactionStatus := constant.StatusPending
	feeDecimal := decimal.NewFromFloat(1000)
	disbursementIntSvc := serviceMocks.NewIDisbursementInternalService(t)
	validDisbursementWithTransaction := &disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: merchantID, Fee: &feeDecimal, UUID: uuid.NewString()}, TransactionStatus: &transactionStatus}

	disbursementIntSvc.On("ExternalFDS", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	boolCmd := &redis.BoolCmd{}
	boolCmd.SetVal(true)
	rdb.On("SetNX", mock.Anything, fmt.Sprintf(constant.DisbursementProcessQueueLockFmt, disbursementID), true, time.Duration(0)).Return(boolCmd)
	rdb.On("Get", mock.Anything, mock.Anything).Return(&redis.StringCmd{})
	rdb.On("Del", mock.Anything, mock.Anything).Return(&redis.IntCmd{})

	testCases := []struct {
		name       string
		mocksSetup func(m *mocker)
		wantErr    bool
	}{
		{
			name: "ERROR: Failed disbursement not found",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: FindByID error",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Merchant not match",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString()}}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Empty transaction status",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: merchantID}}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Failed to acquire mutex lock",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID", mock.Anything, mock.Anything,
				).Return(validDisbursementWithTransaction, nil)

				m.rdb.On(
					"NewMutex", "backend-portal:payouts:"+validDisbursementWithTransaction.UUID+":bank-transfer:lock", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(mutexLock)
				mutexLock.On("LockContext", mock.Anything).Once().Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Failed to retry because processor status not in failed status",
			mocksSetup: func(m *mocker) {
				m.rdb.On(
					"NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(mutexLock)
				mutexLock.On("LockContext", mock.Anything).Return(nil)
				mutexLock.On("UnlockContext", mock.Anything).Once().Return(false, assert.AnError)

				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(validDisbursementWithTransaction, nil)

				m.accountTransactionRepo.On(
					"FindByReference",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: constant.StatusPending}, nil)

				m.snapCoreRepo.On(
					"CheckAllowedToRetry",
					mock.Anything,
					mock.Anything,
				).Return(&snapCoreModel.CheckAllowedToRetryResponse{Allowed: false, Reason: "processor status not in failed status"}, nil)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Retry process",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(validDisbursementWithTransaction, nil)

				mutexLock.On("UnlockContext", mock.Anything).Return(true, nil)

				m.accountTransactionRepo.On(
					"FindByReference",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: constant.StatusPending}, nil)

				m.snapCoreRepo.On(
					"CheckAllowedToRetry",
					mock.Anything,
					mock.Anything,
				).Return(&snapCoreModel.CheckAllowedToRetryResponse{Allowed: true, Reason: ""}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: constant.StatusPending}, nil)

				m.routingProcessorSvc.On(
					"GetTransferByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: constant.SnapCoreBankTransferStatusFailed}, nil)

				m.forbiddenUsecaseSvc.On(
					"CheckUseCase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				m.bankAccountRepo.Mock.On(
					"GetByMerchantID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				m.routingProcessorSvc.On(
					"BankTransfer",
					mock.Anything,
					BankTransferReqMockType,
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: constant.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On(
					"UpdateProcessorReferenceIdAndBankReferenceNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateTransactionTimestamp",
					mock.Anything,
					constant.StringMockType(),
					constant.TimeMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Retry process for flip pg processor",
			mocksSetup: func(m *mocker) {
				processor := constant.FlipPGProcessor
				flipDisbursementWithTransaction := validDisbursementWithTransaction
				flipDisbursementWithTransaction.ProcessorReferenceName = &processor
				flipDisbursementWithTransaction.TotalAmount = decimal.NewFromInt(1000)
				spew.Dump(flipDisbursementWithTransaction)

				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(flipDisbursementWithTransaction, nil)

				m.accountTransactionRepo.On(
					"FindByReference",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: constant.StatusPending, ProcessorReference: constant.FlipPGProcessor, Debit: 1000}, nil)

				m.snapCoreRepo.On(
					"CheckAllowedToRetry",
					mock.Anything,
					mock.Anything,
				).Return(&snapCoreModel.CheckAllowedToRetryResponse{Allowed: true, Reason: ""}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: constant.StatusPending, ProcessorReference: constant.FlipPGProcessor, Debit: 1000}, nil)

				m.routingProcessorSvc.On(
					"GetTransferByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: constant.SnapCoreBankTransferStatusFailed, ProcessorReference: constant.FlipPGProcessor}, nil)

				m.routingProcessorSvc.On(
					"GetFlipEscrowBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&routingProcessorModelEscrowBalance.EscrowBalanceResponse{
					ResponseCode:       "2001800",
					ResponseMessage:    "Successful",
					ProcessorReference: constant.FlipPGProcessor,
					BalanceAmount:      1000000,
				}, nil)

				m.forbiddenUsecaseSvc.On(
					"CheckUseCase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				m.bankAccountRepo.Mock.On(
					"GetByMerchantID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				m.routingProcessorSvc.On(
					"BankTransfer",
					mock.Anything,
					BankTransferReqMockType,
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: constant.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On(
					"UpdateProcessorReferenceIdAndBankReferenceNo",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateTransactionTimestamp",
					mock.Anything,
					constant.StringMockType(),
					constant.TimeMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)

			m := &mocker{
				disbursementRepo:               repositoryMocks.NewIDisbursementRepository(t),
				snapCoreRepo:                   repositoryMocks.NewISnapCoreRepository(t),
				bankAccountRepo:                repositoryMocks.NewIBankAccountRepository(t),
				accountTransactionRepo:         repositoryMocks.NewIAccountTransactionRepository(t),
				orchestratorSvc:                serviceMocks.NewIOrchestratorService(t),
				beneficiaryAccSvc:              serviceMocks.NewIBeneficiaryAccountService(t),
				rmqExt:                         rabbitMqMocks.NewRabbitMQExt(t),
				forbiddenUsecaseSvc:            serviceMocks.NewIMerchantForbiddenUseCaseService(t),
				feeSvc:                         serviceMocks.NewIFeeService(t),
				routingProcessorSvc:            serviceMocks.NewIRoutingProcessorService(t),
				statusHistoriesRepo:            repositoryMocks.NewIStatusHistoriesRepository(t),
				payoutManualProcessingAcctRepo: repositoryMocks.NewIPayoutManualProcessingAccountRepository(t),
				rdb:                            rdb,
			}
			m.statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()
			m.payoutManualProcessingAcctRepo.On(
				"IsManualProcessingAccount",
				mock.Anything,
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
			).Return(false, nil).Maybe()

			tc.mocksSetup(m)

			svc := New(
				&conf, pdkLoggerMock, merchantRepo, m.disbursementRepo, m.snapCoreRepo, m.bankAccountRepo,
				WithOrchestratorService(m.orchestratorSvc),
				WithBeneficiaryAccService(m.beneficiaryAccSvc),
				WithMerchantForbiddenUseCaseService(m.forbiddenUsecaseSvc),
				// WithRedisClient(redisExt.WrapRedisClient(db, nil)),
				WithRedisClient(rdb),
				WithFeeService(m.feeSvc),
				WithRoutingProcessorService(m.routingProcessorSvc),
				WithStatusHistoriesRepository(m.statusHistoriesRepo),
				WithPayoutManualProcessingAccountRepository(m.payoutManualProcessingAcctRepo),
				WithDisbursementInternalService(disbursementIntSvc),
				WithAccountTransactionRepository(m.accountTransactionRepo),
			)
			ctx := context.Background()
			err := svc.RetryDueToInsufficientEscrowFund(ctx, &disbursementModel.RetryTransaction{
				DisbursementID: disbursementID,
				MerchantID:     merchantID,
			})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
