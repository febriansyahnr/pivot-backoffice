package disbursementService

import (
	"bytes"
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApprove(t *testing.T) {
	type mocker struct {
		disbursementRepo    *repositoryMocks.IDisbursementRepository
		orchestratorSvc     *serviceMocks.IOrchestratorService
		rmqExt              *rabbitMqMocks.RabbitMQExt
		internal            *serviceMocks.IDisbursementInternalService
		redisExt            *redisExtMocks.IRedisExt
		redisMutex          *redisExtMocks.IMutexer
		statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository
	}

	disbursementID := uuid.NewString()

	validInput := &disbursementModel.ApproveRequest{
		ApproveAction: []disbursementModel.ApproveActionObject{
			{
				DisbursementID: disbursementID,
			},
		},
		MerchantID: uuid.NewString(),
		ApprovedBy: uuid.NewString(),
	}

	validInputBulk := &disbursementModel.ApproveRequest{
		BulkID: uuid.NewString(),
		ApproveAction: []disbursementModel.ApproveActionObject{
			{
				DisbursementID: disbursementID,
			},
			{
				DisbursementID: uuid.NewString(),
			},
		},
		MerchantID: uuid.NewString(),
		ApprovedBy: uuid.NewString(),
	}

	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	ctxValue := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, "ABC")
	feeDecimal := decimal.NewFromFloat(1000)

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	defer buf.Reset()

	logger := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	testCases := []struct {
		name       string
		mocksSetup func(m *mocker)
		input      *disbursementModel.ApproveRequest
		wantErr    bool
	}{
		{
			name: "SUCCESS:Approval For Bulk Transaction",
			mocksSetup: func(m *mocker) {
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Once().Return(true, nil)
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxValue, nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)

				m.disbursementRepo.On(
					"ApproveInBulk", mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(100000.00, nil)

				m.disbursementRepo.On(
					"SumAmountByIDs", mock.Anything, constant.ArrayStringMockType(),
				).Once().Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)

				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.disbursementRepo.On(
					"FindByID", mock.Anything, constant.StringMockType(),
				).Times(2).Return(&disbursementModel.DisbursementWithTransaction{Disbursement: disbursementModel.Disbursement{MerchantID: uuid.NewString(), Fee: &feeDecimal}}, nil)

				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Twice().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Twice().Return(true, nil)
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Twice().Return(m.redisMutex, nil)

				m.internal.On(
					"CreatePendingOrchestratorTransaction", mock.Anything, mock.Anything,
				).Times(2).Return("1", "2", nil)
				m.internal.On(
					"ValidateBankAccountAndUpdateTransaction", constant.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Times(2).Return(nil)
				m.rmqExt.On(
					"Publish", constant.ValueCtxMockType(), constant.StringMockType(), mock.Anything, mock.Anything,
				).Once().Return(nil)
				m.internal.On(
					"GetCutOffTimeStatus", constant.ValueCtxMockType(), constant.TimeMockType(), constant.StringMockType(), mock.Anything,
				).Once().Return(&disbursementModel.CutOffTimeStatusResponse{
					Status: constant.DisbursementCutOffTimeStatusOngoing,
				}, nil)
				m.redisExt.On("Expire", constant.ValueCtxMockType(), constant.DelayTransferProcessRedisKey, constant.DurationMockType()).Return(nil)
				m.redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), constant.DelayTransferProcessRedisKey, "total", 1.00).Times(2).Return(1.00, nil)
				m.redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), constant.DelayTransferProcessRedisKey, "amount", constant.Float64MockType()).Times(2).Return(0.00, nil)
				m.redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), constant.DelayTransferProcessRedisKey, "bank__total", constant.Float64MockType()).Times(2).Return(0.00, nil)
				m.redisExt.On("HIncrByFloat", constant.ValueCtxMockType(), constant.DelayTransferProcessRedisKey, "bank__amount", constant.Float64MockType()).Times(2).Return(0.00, nil)
			},
			input: validInputBulk, wantErr: false,
		},
		{
			name:       "success:Empty Approval List",
			mocksSetup: func(m *mocker) { /* Empty Function */ },
			input:      &disbursementModel.ApproveRequest{},
			wantErr:    false,
		},
		{
			name: "FAILED:Exclusive Lock",
			mocksSetup: func(m *mocker) {
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			input: validInput, wantErr: true,
		},
		{
			name: "FAILED:Begin Transaction",
			mocksSetup: func(m *mocker) {
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Once().Return(false, constant.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			input: validInput, wantErr: true,
		},
		{
			name: "ERROR:Approve In Bulk Transaction",
			mocksSetup: func(m *mocker) {
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Once().Return(true, nil)

				m.disbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Once().Return(ctxValue, nil)
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On(
					"ApproveInBulk", mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			input: validInput, wantErr: true,
		},
		{
			name: "ERROR:Approve In Bulk Transaction No Rows Affected",
			mocksSetup: func(m *mocker) {
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Once().Return(true, nil)

				m.disbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Once().Return(ctxValue, nil)
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)

				m.disbursementRepo.On(
					"ApproveInBulk", mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrNoRowsAffected)
			},
			input: validInput, wantErr: true,
		},
		{
			name: "ERROR:Approve Action Insufficient Balance/Failed Commit",
			mocksSetup: func(m *mocker) {
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Once().Return(true, nil)

				m.disbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Once().Return(ctxValue, nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).Once().Return(nil)

				m.disbursementRepo.On(
					"ApproveInBulk", mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(0.0, nil)

				m.disbursementRepo.On(
					"SumAmountByIDs", mock.Anything, constant.ArrayStringMockType(),
				).Once().Return(&disbursementModel.SumAmountResponse{TotalAmount: 10000}, nil)

				m.disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)
			},
			input: validInput, wantErr: true,
		},
		{
			name: "ERROR:Approve Action Insufficient Balance",
			mocksSetup: func(m *mocker) {
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Once().Return(true, nil)

				m.disbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Once().Return(ctxValue, nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)

				m.disbursementRepo.On(
					"ApproveInBulk", mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(0.0, nil)

				m.disbursementRepo.On(
					"SumAmountByIDs", mock.Anything, constant.ArrayStringMockType(),
				).Once().Return(&disbursementModel.SumAmountResponse{TotalAmount: 10000}, nil)

				m.disbursementRepo.On(
					"UpdateReasonByIDs",
					mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)
				m.internal.On(
					"DecrDailyTransactionLimit", constant.ValueCtxMockType(), constant.StringMockType(), constant.Float64MockType(),
				).Once().Return(nil)
			},
			input: validInput, wantErr: true,
		},
		{
			name: "ERROR:Process Single Transaction/Failed Commit",
			mocksSetup: func(m *mocker) {
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Once().Return(true, nil)

				m.disbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Once().Return(ctxValue, nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).Once().Return(nil)

				m.disbursementRepo.On(
					"ApproveInBulk", mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(11_000.0, nil)

				m.disbursementRepo.On(
					"SumAmountByIDs", mock.Anything, constant.ArrayStringMockType(),
				).Once().Return(&disbursementModel.SumAmountResponse{TotalAmount: 10_000}, nil)
				m.internal.On(
					"GetCutOffTimeStatus", constant.ValueCtxMockType(), constant.TimeMockType(), constant.StringMockType(), mock.Anything,
				).Once().Return(&disbursementModel.CutOffTimeStatusResponse{}, nil)
			},
			input: validInput, wantErr: true,
		},
		{
			name: "ERROR:Update Bulk Disbursement Status",
			mocksSetup: func(m *mocker) {
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Once().Return(true, nil)
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxValue, nil)
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).Once().Return(nil)

				m.disbursementRepo.On(
					"ApproveInBulk", mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(100000.00, nil)

				m.disbursementRepo.On(
					"SumAmountByIDs", mock.Anything, constant.ArrayStringMockType(),
				).Once().Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)
				m.internal.On(
					"GetCutOffTimeStatus", constant.ValueCtxMockType(), constant.TimeMockType(), constant.StringMockType(), mock.Anything,
				).Once().Return(&disbursementModel.CutOffTimeStatusResponse{}, nil)
				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			input: validInputBulk, wantErr: true,
		},
		{
			name: "ERROR:Approval Bulk Transaction/Failed Commit",
			mocksSetup: func(m *mocker) {
				m.redisMutex.On("LockContext", constant.ValueCtxMockType()).Once().Return(nil)
				m.redisMutex.On("UnlockContext", mock.Anything).Once().Return(true, nil)
				m.redisExt.On(
					"NewMutex", constant.StringMockType(), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(m.redisMutex, nil)

				m.disbursementRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxValue, nil)
				m.disbursementRepo.On("CommitTransaction", mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				m.disbursementRepo.On("RollbackTransaction", mock.Anything).Once().Return(nil)

				m.disbursementRepo.On(
					"ApproveInBulk", mock.Anything, constant.ArrayStringMockType(), constant.StringMockType(),
				).Once().Return(nil)

				m.orchestratorSvc.On(
					"GetAvailableMerchantBalance", mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(100000.00, nil)

				m.disbursementRepo.On(
					"SumAmountByIDs", mock.Anything, constant.ArrayStringMockType(),
				).Once().Return(&disbursementModel.SumAmountResponse{TotalAmount: 0}, nil)
				m.internal.On(
					"GetCutOffTimeStatus", constant.ValueCtxMockType(), constant.TimeMockType(), constant.StringMockType(), mock.Anything,
				).Once().Return(&disbursementModel.CutOffTimeStatusResponse{}, nil)
				m.disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)
			},
			input: validInputBulk, wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			m := &mocker{
				disbursementRepo:    repositoryMocks.NewIDisbursementRepository(t),
				orchestratorSvc:     serviceMocks.NewIOrchestratorService(t),
				rmqExt:              rabbitMqMocks.NewRabbitMQExt(t),
				internal:            serviceMocks.NewIDisbursementInternalService(t),
				redisExt:            redisExtMocks.NewIRedisExt(t),
				redisMutex:          redisExtMocks.NewIMutexer(t),
				statusHistoriesRepo: repositoryMocks.NewIStatusHistoriesRepository(t),
			}

			// General status history mock
			m.statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			tc.mocksSetup(m)

			svc := New(
				&conf, logger, nil, m.disbursementRepo, nil, nil,
				WithStatusHistoriesRepository(m.statusHistoriesRepo),
				WithRabbitMQClient(m.rmqExt),
				WithOrchestratorService(m.orchestratorSvc),
				WithRedisClient(m.redisExt),
				WithDisbursementInternalService(m.internal),
			)

			if err := svc.Approve(context.Background(), tc.input); tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
