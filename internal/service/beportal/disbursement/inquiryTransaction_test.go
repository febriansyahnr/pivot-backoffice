package disbursementService

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInquiryTransaction(t *testing.T) {
	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}
	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
	routingProcessorSvc := serviceMocks.NewIRoutingProcessorService(t)
	orchSvc := serviceMocks.NewIOrchestratorService(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	feeSvc := serviceMocks.NewIFeeService(t)

	// General status history mock that will handle any calls
	statusHistoriesRepo.On("GetByReference", mock.Anything, mock.Anything, mock.Anything).Return([]*statusHistoryModel.StatusHistory{}, nil).Maybe()

	validDisbursementID := uuid.NewString()
	validMerchantID := uuid.NewString()
	validBulkID := uuid.NewString()

	pendingStatus := constant.StatusPending
	successStatus := constant.StatusSuccess

	tests := []struct {
		name         string
		modifierMock func()
		wantErr      bool
		want         *disbursementModel.DisbursementWithTransaction
	}{
		{
			name: "ERROR: FindByID not found",
			modifierMock: func() {
				disbursementRepo.On(
					"FindByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Merchant is not match",
			modifierMock: func() {
				disbursementRepo.On(
					"FindByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Once().Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						MerchantID: "not match",
					},
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Transaction is not created yet",
			modifierMock: func() {
				disbursementRepo.On(
					"FindByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Once().Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						MerchantID: validMerchantID,
					},
					TransactionStatus: nil,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Transaction not in pending status",
			modifierMock: func() {
				disbursementRepo.On(
					"FindByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Once().Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						MerchantID: validMerchantID,
					},
					TransactionStatus: &successStatus,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Inquiry transaction",
			modifierMock: func() {
				disbursementRepo.On(
					"FindByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						MerchantID: validMerchantID,
						BulkID:     &validBulkID,
					},
					TransactionStatus: &pendingStatus,
				}, nil)

				orchSvc.On(
					"FindByReference",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS",
			modifierMock: func() {
				orchSvc.On(
					"FindByReference",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{
					Status: constant.StatusPending,
					ReasonType: sql.NullString{
						Valid:  true,
						String: constant.ReasonTypePayoutDelayed,
					},
					ReasonDescription: sql.NullString{
						Valid:  true,
						String: constant.ReasonDescPayoutDelayed,
					},
				}, nil)

				routingProcessorSvc.On(
					"GetTransferByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&routingProcessorModel.BankTransferResponseData{
					Status: constant.StatusPending,
				}, nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				orchSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
				).Return(nil)
				disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				disbursementRepo.On(
					"CountStatusInProgressByBulkID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(0)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.modifierMock()

			svc := New(conf, logger, nil, disbursementRepo, snapCoreRepo, nil, WithStatusHistoriesRepository(statusHistoriesRepo), WithOrchestratorService(orchSvc), WithFeeService(feeSvc), WithRoutingProcessorService(routingProcessorSvc))

			_, err := svc.InquiryTransaction(context.Background(), &disbursementModel.InquiryTransaction{
				DisbursementID: validDisbursementID,
				MerchantID:     validMerchantID,
			})
			if !tc.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestRetryInquirePendingTransactions(t *testing.T) {
	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 5,
		},
	}
	redisExt := redisMock.NewIRedisExt(t)
	mutexLock := redisMock.NewIMutexer(t)
	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
	routingProcessorSvc := serviceMocks.NewIRoutingProcessorService(t)
	orchSvc := serviceMocks.NewIOrchestratorService(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	feeSvc := serviceMocks.NewIFeeService(t)
	disbursementIntSvc := serviceMocks.NewIDisbursementInternalService(t)

	// General status history mock that will handle any calls
	statusHistoriesRepo.On("GetByReference", mock.Anything, mock.Anything, mock.Anything).Return([]*statusHistoryModel.StatusHistory{}, nil).Maybe()

	disbursementAmount := decimal.New(12000, 10)

	svc := New(conf, logger, nil, disbursementRepo, snapCoreRepo, nil,
		WithStatusHistoriesRepository(statusHistoriesRepo),
		WithOrchestratorService(orchSvc),
		WithFeeService(feeSvc),
		WithRoutingProcessorService(routingProcessorSvc),
		WithDisbursementInternalService(disbursementIntSvc),
		WithRedisClient(redisExt),
	)

	tests := []struct {
		name      string
		start     time.Time
		end       time.Time
		setupMock func()
		shouldErr bool
		wantErr   error
		want      *disbursementModel.RetryInquireDisbuesementSummary
	}{
		{
			name:  "ERROR: When failed to GetPendingTransactionsBetweenTime, then should return error",
			start: time.Now().Add(-24 * time.Hour),
			end:   time.Now(),
			setupMock: func() {
				disbursementRepo.On(
					"GetPendingTransactionsBetweenTimeForInquiryTransaction",
					mock.Anything,
					constant.TimeMockType(),
					constant.TimeMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr: true,
			wantErr:   constant.ErrSomeErrorForUnitTest,
		},
		{
			name:  "ERROR: acquire mutex lock",
			start: time.Now().Add(-24 * time.Hour),
			end:   time.Now(),
			setupMock: func() {
				payoutID := uuid.NewString()
				disbursementRepo.On(
					"GetPendingTransactionsBetweenTimeForInquiryTransaction",
					mock.Anything,
					constant.TimeMockType(),
					constant.TimeMockType(),
				).Once().Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:   payoutID,
							Amount: disbursementAmount,
						},
						TransactionStatus: util.ValueToPtr(constant.StatusPending),
					},
				}, nil)
				redisExt.On(
					"NewMutex", "backend-portal:payouts:"+payoutID+":bank-transfer:lock", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(mutexLock)
				mutexLock.On("LockContext", mock.Anything).Once().Return(assert.AnError)
			},
			shouldErr: false,
			want: &disbursementModel.RetryInquireDisbuesementSummary{
				Total:  1,
				Amount: disbursementAmount.InexactFloat64(),
			},
		},
		{
			name:  "ERROR: ProcessInquiryTransaction fails",
			start: time.Now().Add(-24 * time.Hour),
			end:   time.Now(),
			setupMock: func() {
				disbursements := []*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:   uuid.NewString(),
							Amount: disbursementAmount,
						},
						TransactionStatus: util.ValueToPtr(constant.StatusPending),
					},
				}
				disbursementRepo.On(
					"GetPendingTransactionsBetweenTimeForInquiryTransaction",
					mock.Anything,
					constant.TimeMockType(),
					constant.TimeMockType(),
				).Return(disbursements, nil).Once()

				redisExt.On(
					"NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(mutexLock)
				mutexLock.On("LockContext", mock.Anything).Return(nil)
				mutexLock.On("UnlockContext", mock.Anything).Return(false, assert.AnError)

				orchSvc.On(
					"FindByReference",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr: false,
			want: &disbursementModel.RetryInquireDisbuesementSummary{
				Total:        1,
				Amount:       disbursementAmount.InexactFloat64(),
				TotalFailed:  1,
				AmountFailed: disbursementAmount.InexactFloat64(),
			},
		},
		{
			name:  "SUCCESS: ProcessInquiryTransaction succeeds",
			start: time.Now().Add(-24 * time.Hour),
			end:   time.Now(),
			setupMock: func() {
				disbursements := []*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:   uuid.NewString(),
							Amount: disbursementAmount,
						},
						TransactionStatus: util.ValueToPtr(constant.StatusPending),
					},
				}
				disbursementRepo.On(
					"GetPendingTransactionsBetweenTimeForInquiryTransaction",
					mock.Anything,
					constant.TimeMockType(),
					constant.TimeMockType(),
				).Return(disbursements, nil)

				mutexLock.On("UnlockContext", mock.Anything).Return(true, nil)

				orchSvc.On(
					"FindByReference",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.TypeDisbursement,
				).Return(&orchestratorModel.AccountTransactionWithUseCase{
					Status: constant.StatusPending,
				}, nil).Once()
				orchSvc.On(
					"FindByReference",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.TypeFee,
				).Return(&orchestratorModel.AccountTransactionWithUseCase{
					Status: constant.StatusPending,
				}, nil).Once()

				routingProcessorSvc.On(
					"GetTransferByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&routingProcessorModel.BankTransferResponseData{
					Status: constant.StatusPending,
				}, nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil).Once()

				orchSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
				).Return(nil)

				disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil).Once()

				disbursementRepo.On(
					"FindByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:   uuid.NewString(),
						Amount: disbursementAmount,
					},
					TransactionStatus: util.ValueToPtr(string(constant.StatusSuccess)),
				}, nil).Once()

				statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()
			},
			shouldErr: false,
			want: &disbursementModel.RetryInquireDisbuesementSummary{
				Total:           1,
				Amount:          disbursementAmount.InexactFloat64(),
				TotalSucceeded:  1,
				AmountSucceeded: disbursementAmount.InexactFloat64(),
				TotalFailed:     0,
				AmountFailed:    0,
			},
		},
		{
			name:  "SUCCESS: ProcessInquiryTransaction succeeds for FAILED transaction",
			start: time.Now().Add(-24 * time.Hour),
			end:   time.Now(),
			setupMock: func() {
				disbursements := []*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:   uuid.NewString(),
							Amount: disbursementAmount,
						},
						TransactionStatus: util.ValueToPtr(constant.StatusPending),
					},
				}
				disbursementRepo.On(
					"GetPendingTransactionsBetweenTimeForInquiryTransaction",
					mock.Anything,
					constant.TimeMockType(),
					constant.TimeMockType(),
				).Return(disbursements, nil)

				orchSvc.On(
					"FindByReference",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.TypeDisbursement,
				).Return(&orchestratorModel.AccountTransactionWithUseCase{
					Status: constant.StatusPending,
				}, nil).Once()
				orchSvc.On(
					"FindByReference",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.TypeFee,
				).Return(&orchestratorModel.AccountTransactionWithUseCase{
					Status: constant.StatusPending,
				}, nil).Once()

				routingProcessorSvc.On(
					"GetTransferByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&routingProcessorModel.BankTransferResponseData{
					Status: constant.StatusPending,
				}, nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil).Once()

				orchSvc.On(
					"UpdateStatusAccountTransaction",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.PtrStringMockType(),
					constant.PtrStringMockType(),
				).Return(nil)

				disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil).Once()

				disbursementRepo.On(
					"FindByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:                 uuid.NewString(),
						Amount:               disbursementAmount,
						BeneficiaryBankCode:  "002",
						BeneficiaryAccountNo: "123456",
					},
					TransactionStatus: util.ValueToPtr(constant.StatusFailed),
				}, nil).Once()

				statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				disbursementIntSvc.On(
					"DecrBeneficiaryPayoutLimit", mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.Float64MockType(),
				).Return(nil)
			},
			shouldErr: false,
			want: &disbursementModel.RetryInquireDisbuesementSummary{
				Total:           1,
				Amount:          disbursementAmount.InexactFloat64(),
				TotalSucceeded:  0,
				AmountSucceeded: 0,
				TotalFailed:     1,
				AmountFailed:    disbursementAmount.InexactFloat64(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			summary, err := svc.RetryInquirePendingTransactions(context.Background(), tc.start, tc.end)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, summary)
			assert.Equal(t, tc.want, summary)
		})
	}
}
