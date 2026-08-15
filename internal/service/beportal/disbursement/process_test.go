package disbursementService

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcess(t *testing.T) {
	conf := config.Config{
		Environment: c.EnvironmentStaging,
	}
	validRemark := "this is valid remark"

	disbursementID := uuid.NewString()
	feeDecimal := decimal.NewFromFloat(1000)
	validDisbursementWithTransaction := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, UUID: uuid.NewString()},
	}
	queueKey := fmt.Sprintf(c.DisbursementProcessQueueLockFmt, disbursementID)

	// Create a mock logger
	pdkLoggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	disbursementIntSvc := serviceMocks.NewIDisbursementInternalService(t)
	disbursementIntSvc.On(
		"DecrDailyTransactionLimit", mock.Anything, c.StringMockType(), c.Float64MockType(),
	).Return(nil).Maybe()
	disbursementIntSvc.On("ExternalFDS", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	type mocker struct {
		disbursementRepo               *repositoryMocks.IDisbursementRepository
		snapCoreRepo                   *repositoryMocks.ISnapCoreRepository
		bankAccountRepo                *repositoryMocks.IBankAccountRepository
		routingProcessorSvc            *serviceMocks.IRoutingProcessorService
		orchestratorSvc                *serviceMocks.IOrchestratorService
		beneficiaryAccSvc              *serviceMocks.IBeneficiaryAccountService
		forbiddenUsecaseSvc            *serviceMocks.IMerchantForbiddenUseCaseService
		feeSvc                         *serviceMocks.IFeeService
		r                              redismock.ClientMock
		statusHistoriesRepo            *repositoryMocks.IStatusHistoriesRepository
		payoutManualProcessingAcctRepo *repositoryMocks.IPayoutManualProcessingAccountRepository
	}

	testCases := []struct {
		name       string
		mocksSetup func(m *mocker)
		wantErr    bool
	}{
		{
			name: "SUCCESS: Process single disbursement",
			mocksSetup: func(m *mocker) {
				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				// Mock the whitelist check - use MatchExpectationsInOrder(false) to be flexible with key matching
				m.r.MatchExpectationsInOrder(false)
				m.r.ExpectGet("backend-portal:trx-merchant-whitelist:").SetErr(errors.New("redis: nil"))

				m.r.ExpectDel(queueKey).SetVal(1)

				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, Remark: &validRemark}}, nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On(
					"UpdateProcessorReferenceIdAndBankReferenceNo",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateTransactionTimestamp",
					mock.Anything,
					c.StringMockType(),
					c.TimeMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Manual processing account exists for merchant - set pending waiting manual action",
			mocksSetup: func(m *mocker) {
				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.r.MatchExpectationsInOrder(false)
				m.r.ExpectGet("backend-portal:trx-merchant-whitelist:").SetErr(errors.New("redis: nil"))

				m.r.ExpectDel(queueKey).SetVal(1)

				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, Remark: &validRemark}}, nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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

				// Manual processing account exists -> skip BankTransfer, set PENDING + waiting manual action
				m.payoutManualProcessingAcctRepo.On(
					"IsManualProcessingAccount",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(p *string) bool { return p != nil && *p == c.ReasonTypeWaitingManualAction }),
					mock.MatchedBy(func(p *string) bool { return p != nil && *p == c.ReasonDescWaitingManualAction }),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Got error in IsManualProcessingAccount",
			mocksSetup: func(m *mocker) {
				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.r.MatchExpectationsInOrder(false)
				m.r.ExpectGet("backend-portal:trx-merchant-whitelist:").SetErr(errors.New("redis: nil"))

				m.r.ExpectDel(queueKey).SetVal(1)

				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, Remark: &validRemark}}, nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				// IsManualProcessingAccount returns error -> fail early
				m.payoutManualProcessingAcctRepo.On(
					"IsManualProcessingAccount",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Disbursement already processed in snap-core (SUCCESS status)",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, Remark: &validRemark}}, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything, c.StringMockType(), c.TypeDisbursement,
				).Times(1).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: c.StatusPending}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything, c.StringMockType(), c.TypeFee,
				).Times(1).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:   uuid.New(),
					Status: c.StatusPending,
					AdditionalInfo: types.NullJSONText{
						JSONText: []byte(`{"deductionType":"MANUAL"}`),
						Valid:    true,
					},
				}, nil)

				m.routingProcessorSvc.On(
					"GetTransferByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(nil)

				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Disbursement already processed in snap-core (SUCCESS status) with LADDER tiering counter",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, Remark: &validRemark}}, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything, c.StringMockType(), c.TypeDisbursement,
				).Times(1).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: c.StatusPending}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything, c.StringMockType(), c.TypeFee,
				).Times(1).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:   uuid.New(),
					Status: c.StatusPending,
					AdditionalInfo: types.NullJSONText{
						JSONText: []byte(`{"deductionType":"MANUAL","ladderCounterKey":"backend-portal:merchant-fee-counter:fee-uuid:2026-03","ladderCounterIncrement":500000}`),
						Valid:    true,
					},
				}, nil)

				m.routingProcessorSvc.On(
					"GetTransferByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(nil)

				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				m.feeSvc.On("IncrementLadderCounter", mock.Anything, "backend-portal:merchant-fee-counter:fee-uuid:2026-03", int64(500_000)).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Disbursement already processed in snap-core (FAILED Insufficient Fund status)",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, Remark: &validRemark}}, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: c.StatusPending}, nil)

				m.routingProcessorSvc.On(
					"GetTransferByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusFailed, ResponseCode: "4031714"}, nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(nil)

				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Disbursement already processed in orchestrator transaction (status SUCCESS)",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, Remark: &validRemark}}, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: c.StatusSuccess}, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Disbursement not processed yet but already created in orchestrator transaction (status PENDING)",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, Remark: &validRemark}}, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New(), Status: c.StatusPending}, nil)

				m.routingProcessorSvc.On(
					"GetTransferByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)

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
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On(
					"UpdateProcessorReferenceIdAndBankReferenceNo",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateTransactionTimestamp",
					mock.Anything,
					c.StringMockType(),
					c.TimeMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Got error in FindByID",
			mocksSetup: func(m *mocker) {
				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Disbursement not found",
			mocksSetup: func(m *mocker) {
				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Parsing merchant ID",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: "invalid merchantID", Fee: &feeDecimal}}, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in PostAccountTransaction",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(c.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in call BankTransfer",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				disbursementIntSvc.On(
					"DecrBeneficiaryPayoutLimit", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.Float64MockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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
				).Return(nil, c.ErrSomeErrorForUnitTest)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in call BankTransfer internal server error",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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
				).Return(&routingProcessorModel.BankTransferResponseData{
					UUID:         uuid.NewString(),
					Status:       c.SnapCoreBankTransferStatusFailed,
					ResponseCode: "5001800",
				},
					c.ErrSomeErrorForUnitTest)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in call BankTransfer and UpdateStatusAccountTransaction",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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
				).Return(nil, c.ErrSomeErrorForUnitTest)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in BeginTransaction",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in CommitTransaction",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(c.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On("RollbackTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in RollbackTransaction",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(c.ErrSomeErrorForUnitTest)
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).
					Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in UpdateProcessorReferenceIdAndBankReferenceNo",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On(
					"UpdateProcessorReferenceIdAndBankReferenceNo",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				m.disbursementRepo.On("RollbackTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in UpdateProcessorReferenceIdAndBankReferenceNo and RollbackTransaction",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On(
					"UpdateProcessorReferenceIdAndBankReferenceNo",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				m.disbursementRepo.On("RollbackTransaction", mock.Anything).
					Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in UpdateStatusAccountTransaction",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

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
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On(
					"UpdateProcessorReferenceIdAndBankReferenceNo",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateTransactionTimestamp",
					mock.Anything,
					c.StringMockType(),
					c.TimeMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				m.disbursementRepo.On("RollbackTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in UpdateStatusAccountTransaction and RollbackTransaction",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On(
					"UpdateProcessorReferenceIdAndBankReferenceNo",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateTransactionTimestamp",
					mock.Anything,
					c.StringMockType(),
					c.TimeMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)

				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				m.disbursementRepo.On("RollbackTransaction", mock.Anything).
					Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Got error in UpdateProcessorAndReconReferenceByID and RollbackTransaction",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(validDisbursementWithTransaction, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

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
				).Return(&routingProcessorModel.BankTransferResponseData{UUID: uuid.NewString(), Status: c.SnapCoreBankTransferStatusSuccess}, nil)

				m.disbursementRepo.On(
					"UpdateProcessorReferenceIdAndBankReferenceNo",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil)

				m.orchestratorSvc.On(
					"UpdateProcessorAndReconReferenceByID",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				m.disbursementRepo.On("RollbackTransaction", mock.Anything).
					Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Rejected due to merchant forbidden use case",
			mocksSetup: func(m *mocker) {
				m.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					c.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal, Remark: &validRemark}}, nil)

				m.r.ExpectSetNX(queueKey, true, 0).SetVal(true)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(100000), nil)

				m.disbursementRepo.On(
					"SumAmountByIDs",
					mock.Anything,
					c.ArrayStringMockType(),
				).Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.orchestratorSvc.On(
					"FindByReference",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				m.orchestratorSvc.On(
					"PostAccountTransaction",
					mock.Anything,
					c.PtrCreateAccTransactionReqMockType(),
				).Return(nil)

				m.forbiddenUsecaseSvc.On(
					"CheckUseCase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error"))

				m.orchestratorSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
					c.PtrStringMockType(),
					c.PtrStringMockType(),
				).Return(nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			disbursementRepoMock := repositoryMocks.NewIDisbursementRepository(t)
			snapCoreRepoMock := repositoryMocks.NewISnapCoreRepository(t)
			bankAccountRepoMock := repositoryMocks.NewIBankAccountRepository(t)
			orchSvcMock := serviceMocks.NewIOrchestratorService(t)
			beneficiaryAccSvcMock := serviceMocks.NewIBeneficiaryAccountService(t)
			forbiddenUsecaseSvc := serviceMocks.NewIMerchantForbiddenUseCaseService(t)
			feeSvc := serviceMocks.NewIFeeService(t)
			db, r := redismock.NewClientMock()

			// Generate a predictable reference ID for use in Redis keys
			refId := "test-reference-id"
			validDisbursementWithTransaction.ReferenceID = refId

			// Setup all required Redis mocks with specific keys
			// Only set expectations for the specific test case being run
			// Don't set global expectations here that might conflict with test case expectations

			m := &mocker{
				disbursementRepo:               disbursementRepoMock,
				snapCoreRepo:                   snapCoreRepoMock,
				bankAccountRepo:                bankAccountRepoMock,
				routingProcessorSvc:            serviceMocks.NewIRoutingProcessorService(t),
				orchestratorSvc:                orchSvcMock,
				beneficiaryAccSvc:              beneficiaryAccSvcMock,
				forbiddenUsecaseSvc:            forbiddenUsecaseSvc,
				feeSvc:                         feeSvc,
				r:                              r,
				statusHistoriesRepo:            repositoryMocks.NewIStatusHistoriesRepository(t),
				payoutManualProcessingAcctRepo: repositoryMocks.NewIPayoutManualProcessingAccountRepository(t),
			}
			m.statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			tc.mocksSetup(m)

			// Default: manual processing account does not exist. Cases that need the manual path override this after setup.
			m.payoutManualProcessingAcctRepo.On(
				"IsManualProcessingAccount",
				mock.Anything,
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
			).Return(false, nil).Maybe()

			svc := New(
				&conf, pdkLoggerMock, merchantRepo, m.disbursementRepo, m.snapCoreRepo, m.bankAccountRepo,
				WithOrchestratorService(m.orchestratorSvc),
				WithBeneficiaryAccService(m.beneficiaryAccSvc),
				WithMerchantForbiddenUseCaseService(m.forbiddenUsecaseSvc),
				WithRoutingProcessorService(m.routingProcessorSvc),
				WithRedisClient(redisExt.WrapRedisClient(db, nil)),
				WithFeeService(m.feeSvc),
				WithDisbursementInternalService(disbursementIntSvc),
				WithStatusHistoriesRepository(m.statusHistoriesRepo),
				WithPayoutManualProcessingAccountRepository(m.payoutManualProcessingAcctRepo),
			)

			ctx := context.WithValue(context.Background(), c.CtxParentMerchantId, uuid.NewString())

			err := svc.Process(ctx, disbursementID, false)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			disbursementRepoMock.AssertExpectations(t)
			snapCoreRepoMock.AssertExpectations(t)
			orchSvcMock.AssertExpectations(t)
			beneficiaryAccSvcMock.AssertExpectations(t)
		})
	}
}

func TestCreateBankTransfer(t *testing.T) {
	config := &config.Config{
		Environment: constant.EnvironmentStaging,
	}

	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	disbursementInternalSvc := serviceMocks.NewIDisbursementInternalService(t)

	trxId := "051ee234-a352-49b5-9676-e6688960e144"
	disbursementId := "42790229-c769-4d43-b751-63e1288cc5d1"

	service := &DisbursementService{
		config:          config,
		logger:          logger,
		orchestratorSvc: orchestratorSvc,
		self:            disbursementInternalSvc,
	}

	tests := []struct {
		name                 string
		environment          string
		beneficiaryAccountNo string
		setupMock            func()
		wantErr              error
	}{
		{
			name:                 "ERROR:Find By Disbursement Reference/Transaction",
			beneficiaryAccountNo: "999966660007",
			setupMock: func() {
				orchestratorSvc.On(
					"FindByReference", constant.ValueCtxMockType(), disbursementId, constant.TypeDisbursement,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:        "ERROR:Disbursement Data Not Found/Development",
			environment: "development",
			setupMock: func() {
				orchestratorSvc.On(
					"FindByReference", constant.ValueCtxMockType(), disbursementId, constant.TypeDisbursement,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, constant.ErrLedgerDetailNotFound),
		},
		{
			name:        "ERROR:Disbursement Data Not Found/Staging",
			environment: constant.EnvironmentStaging,
			setupMock: func() {
				orchestratorSvc.On(
					"FindByReference", constant.ValueCtxMockType(), disbursementId, constant.TypeDisbursement,
				).Once().Return(nil, nil)
				disbursementInternalSvc.On(
					"CreatePendingOrchestratorTransaction", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return("", "", constant.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Transaction Already Processed",
			setupMock: func() {
				orchestratorSvc.On(
					"FindByReference", constant.ValueCtxMockType(), disbursementId, constant.TypeDisbursement,
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{Status: constant.StatusSuccess}, nil)
			},
		},
		{
			name:        "ERROR:Find By Disbursement Reference/Fee",
			environment: constant.EnvironmentStaging,
			setupMock: func() {
				orchestratorSvc.On(
					"FindByReference", constant.ValueCtxMockType(), disbursementId, constant.TypeDisbursement,
				).Times(1).Return(nil, nil)
				disbursementInternalSvc.On(
					"CreatePendingOrchestratorTransaction", constant.ValueCtxMockType(), mock.Anything,
				).Return(trxId, trxId, nil)

				orchestratorSvc.On(
					"FindByReference", constant.ValueCtxMockType(), disbursementId, constant.TypeFee,
				).Times(1).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:        "SUCCESS",
			environment: constant.EnvironmentStaging,
			setupMock: func() {
				orchestratorSvc.On(
					"FindByReference", constant.ValueCtxMockType(), disbursementId, constant.StringMockType(),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: util.ParseUUID(trxId), Status: constant.StatusPending}, nil)

				disbursementInternalSvc.On(
					"ProcessBankTransferAndUpdateTransaction", constant.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if test.environment != "" {
				config.Environment = test.environment
			}
			assert.Equal(
				t, test.wantErr, service.CreateBankTransfer(
					context.Background(), &disbursementModel.DisbursementWithTransaction{
						Disbursement: disbursementModel.Disbursement{
							UUID:                 disbursementId,
							BeneficiaryAccountNo: test.beneficiaryAccountNo,
						},
					}))
		})
	}
}

func TestUpdateTransactionStatus(t *testing.T) {
	ledgerSvc := serviceMocks.NewILedgerService(t)
	transferSvc := serviceMocks.NewITransferService(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)

	service := &DisbursementService{
		ledgerSvc:       ledgerSvc,
		transferSvc:     transferSvc,
		orchestratorSvc: orchestratorSvc,
	}

	tests := []struct {
		name        string
		transaction *orchestratorModel.TransactionAndFeeObject
		setupMock   func()
		wantErr     error
	}{
		{
			name:        "ERROR:Update disbursement only",
			transaction: &orchestratorModel.TransactionAndFeeObject{},
			setupMock: func() {
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything, mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name:        "SUCCESS:Update disbursement only",
			transaction: &orchestratorModel.TransactionAndFeeObject{},
			setupMock: func() {
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything, mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name: "ERROR:Update disbursement and fee",
			transaction: &orchestratorModel.TransactionAndFeeObject{
				FeeID: uuid.NewString(),
			},
			setupMock: func() {
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything, mock.Anything,
				).Times(1).Return(nil)

				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything, mock.Anything,
				).Times(1).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Update disbursement and fee",
			transaction: &orchestratorModel.TransactionAndFeeObject{
				FeeID: uuid.NewString(),
			},
			setupMock: func() {
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything, mock.Anything,
				).Return(nil)
			},
		},
		{
			name: "ERROR:Update transaction on behalf",
			transaction: &orchestratorModel.TransactionAndFeeObject{
				FeeID:         uuid.NewString(),
				TransferFeeID: uuid.NewString(),
			},
			setupMock: func() {
				ledgerSvc.On(
					"UpdateTransaction", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Update transaction on behalf",
			transaction: &orchestratorModel.TransactionAndFeeObject{
				FeeID:         uuid.NewString(),
				TransferFeeID: uuid.NewString(),
			},
			setupMock: func() {
				ledgerSvc.On(
					"UpdateTransaction", constant.ValueCtxMockType(), mock.Anything,
				).Return(nil)
				transferSvc.On(
					"UpdateTransferStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything,
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(
				t, test.wantErr, service.updateTransactionStatus(
					context.Background(), test.transaction, constant.StatusFailed, util.ValueToPtr(constant.ReasonTypeOtherReason), util.ValueToPtr(constant.ReasonDescInvalidBeneficiaryAccount),
				),
			)

			ledgerSvc.AssertExpectations(t)
			transferSvc.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
		})
	}
}

func TestValidateBankAccountAndUpdateTransactionSuccess(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	// Mocks
	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	beneficiarySvc := serviceMocks.NewIBeneficiaryAccountService(t)
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	redisExt := redisExtMocks.NewIRedisExt(t)
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)

	// Expectation: beneficiary account is valid; no transaction update
	derivedMerchantID := uuid.NewString()
	beneficiarySvc.On(
		"FindByBankCodeAndAccountNo",
		mock.Anything,
		mock.MatchedBy(func(req *beneficiaryAccountModel.CheckAccountRequest) bool {
			return req != nil && req.MerchantID == derivedMerchantID
		}),
	).Once().Return(&beneficiaryAccountModel.Account{}, nil)
	// Status history insert is optional in this path; allow if called elsewhere
	statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()
	redisExt.On("HGetAllScan", mock.Anything, constant.StringMockType(), mock.Anything).
		Run(func(args mock.Arguments) {
			limitResp := args.Get(2).(*disbursementModel.BeneficiaryPayoutLimitRuleLimit)
			limitResp.Count = 5
		}).
		Return(nil).Once()

	redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), constant.StringMockType(), "processed", mock.Anything).Return(2000.00, nil)
	redisExt.On("HIncrBy", constant.ValueCtxMockType(), constant.StringMockType(), "count", int64(1)).Return(int64(6), nil)

	merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{
		UUID: "other-merchant-id",
		Name: "Test Merchant",
	}, nil)

	svcAny := New(
		&config.Config{Environment: constant.EnvironmentStaging, DisbursementConfig: config.DisbursementConfig{
			BeneficiaryLimit: config.DisbursementBeneficiaryLimit{
				Amount:   10000000.00,
				Velocity: 100,
			},
		}},
		logger,
		merchantRepo, disbursementRepo, nil, nil,
		WithBeneficiaryAccService(beneficiarySvc),
		WithOrchestratorService(orchestratorSvc),
		WithStatusHistoriesRepository(statusHistoriesRepo),
		WithRedisClient(redisExt),
	)
	svc := svcAny.(*DisbursementService)

	disb := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:                   uuid.NewString(),
			MerchantID:             uuid.NewString(),
			BeneficiaryAccountNo:   "1234567890",
			BeneficiaryBankCode:    "008",
			BeneficiaryAccountName: "John Doe",
			CreatedAt:              time.Now().UTC(),
		},
	}

	trx := &orchestratorModel.TransactionAndFeeObject{
		TransactionID: uuid.NewString(),
		FeeID:         "", // ensure no fee updates are required in test
	}

	// Context with derived merchant id
	ctx := context.WithValue(context.Background(), constant.CtxDerivedMerchantID, derivedMerchantID)

	err := svc.ValidateBankAccountAndUpdateTransaction(ctx, disb, trx)
	assert.NoError(t, err)
}

func TestValidateBankAccountAndUpdateTransactionInvalidBeneficiaryVariants(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	type testCase struct {
		name         string
		svcErr       error
		expectedDesc string
	}

	cases := []testCase{
		{
			name:         "ERROR: invalid account maps to invalid message",
			svcErr:       constant.ErrInvalidAccount,
			expectedDesc: constant.SnapCoreResponseInvalidAccountMessage,
		},
		{
			name:         "ERROR: inactive account maps to inactive message",
			svcErr:       constant.ErrInactiveAccount,
			expectedDesc: constant.SnapCoreResponseInactiveAccountMessage,
		},
		{
			name:         "ERROR: dormant account maps to dormant message",
			svcErr:       constant.ErrDormantAccount,
			expectedDesc: constant.SnapCoreResponseDormantAccountMessage,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Mocks
			disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
			orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			beneficiarySvc := serviceMocks.NewIBeneficiaryAccountService(t)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)

			// Begin/Commit transaction expectations
			trxCtx := context.WithValue(context.Background(), "sql_tx", "dummy")
			disbursementRepo.On("BeginTransaction", mock.Anything).Once().Return(trxCtx, nil)
			disbursementRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)

			// Beneficiary check returns specific error
			derivedMerchantID := uuid.NewString()
			beneficiarySvc.On(
				"FindByBankCodeAndAccountNo",
				mock.Anything,
				mock.MatchedBy(func(req *beneficiaryAccountModel.CheckAccountRequest) bool {
					return req != nil && req.MerchantID == derivedMerchantID
				}),
			).Once().Return((*beneficiaryAccountModel.Account)(nil), tc.svcErr)

			// Update transaction status with expected reason type and description
			orchestratorSvc.On(
				"UpdateStatusAccountTransaction",
				mock.Anything,
				constant.StringMockType(),
				constant.StringMockType(),
				mock.MatchedBy(func(p *string) bool { return p != nil && *p == constant.ReasonTypeBeneficiaryAccountReason }),
				mock.MatchedBy(func(p *string) bool { return p != nil && *p == tc.expectedDesc }),
			).Once().Return(nil)

			// Status history insert is non-critical; allow call
			statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			svcAny := New(
				&config.Config{Environment: constant.EnvironmentStaging},
				logger,
				nil, disbursementRepo, nil, nil,
				WithBeneficiaryAccService(beneficiarySvc),
				WithOrchestratorService(orchestratorSvc),
				WithStatusHistoriesRepository(statusHistoriesRepo),
			)
			svc := svcAny.(*DisbursementService)

			disb := &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					UUID:                   uuid.NewString(),
					MerchantID:             uuid.NewString(),
					BeneficiaryAccountNo:   "9876543210",
					BeneficiaryBankCode:    "008",
					BeneficiaryAccountName: "Jane Doe",
					CreatedAt:              time.Now().UTC(),
				},
			}
			trx := &orchestratorModel.TransactionAndFeeObject{
				TransactionID: uuid.NewString(),
				FeeID:         "", // ensure no fee updates are required in test
			}

			ctx := context.WithValue(context.Background(), constant.CtxDerivedMerchantID, derivedMerchantID)
			err := svc.ValidateBankAccountAndUpdateTransaction(ctx, disb, trx)
			// should bubble up the same error
			assert.Error(t, err)
		})
	}
}
