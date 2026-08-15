package disbursementService

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/panjf2000/ants/v2"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessChangeDisbursementTansactionStatus(t *testing.T) {
	ctx := context.Background()
	mockDisbursementRepo := repoMocks.NewIDisbursementRepository(t)
	mockOrchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	mockDisbursementInternalSvc := serviceMocks.NewIDisbursementInternalService(t)
	mockSnapCoreRepo := repoMocks.NewISnapCoreRepository(t)
	mockRedis := redisMocks.NewIRedisExt(t)
	mockStatusHistoriesRepo := repoMocks.NewIStatusHistoriesRepository(t)

	mockTransferSvc := serviceMocks.NewITransferService(t)
	mockFeeSvc := serviceMocks.NewIFeeService(t)
	mockLogger, _ := logger.NewZapLogger(logger.Config{})

	service := DisbursementService{
		logger:              mockLogger,
		disbursementRepo:    mockDisbursementRepo,
		orchestratorSvc:     mockOrchestratorSvc,
		self:                mockDisbursementInternalSvc,
		transferSvc:         mockTransferSvc,
		snapCoreRepo:        mockSnapCoreRepo,
		redisExt:            mockRedis,
		statusHistoriesRepo: mockStatusHistoriesRepo,
		feeSvc:              mockFeeSvc,
	}

	// General status history mock that will handle any calls
	mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

	validUUID, _ := uuid.Parse(constant.EmptyUUID)
	validDisbursementID := "valid-disbursement-id"
	validRequest := disbursementModel.ChangeDisbursementTransactionStatusRequest{
		DisbursementIDS: []string{validDisbursementID},
		Status:          constant.StatusPending,
	}

	disbursementTransaction := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:       validDisbursementID,
			MerchantID: constant.EmptyUUID,
			Amount:     decimal.New(1000, 0),
		},
		TransactionStatus:     util.ValueToPtr(constant.StatusPending),
		TransactionReasonType: util.ValueToPtr(constant.ReasonTypeBlockedByHarsya),
	}

	disbursementLedger := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:               validUUID,
		MerchantID:         validUUID,
		ReferenceID:        validDisbursementID,
		Status:             constant.StatusPending,
		ProcessorReference: constant.FlipPGProcessor,
	}

	testCases := []struct {
		name           string
		disbursementID string
		payload        disbursementModel.ChangeDisbursementTransactionStatusRequest
		setupMock      func()
		shouldError    bool
		wantErr        error
	}{
		{
			name:           "when failed to get disbursement, then should return error",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrSomeErrorForUnitTest,
		},
		{
			name:           "when disbursement not found, then should return error",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(nil, nil).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrDataNotFound,
		},
		{
			name:           "when failed to get disbursement ledger, then should return error",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrSomeErrorForUnitTest,
		},
		{
			name:           "when disbursement ledger not found, then should return error",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(nil, nil).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrDataNotFound,
		},
		{
			name:           "when disbursement ledger status is success, then should return error",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        validUUID,
					MerchantID:  validUUID,
					ReferenceID: validDisbursementID,
					Status:      constant.StatusSuccess,
				}, nil).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrTransactionAlreadySucceeded,
		},
		{
			name:           "when failed to get fee ledger, then should return error",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrSomeErrorForUnitTest,
		},
		{
			name:           "when failed to create sql transaction, then should return error",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrSomeErrorForUnitTest,
		},
		{
			name:           "when failed to update transaction status, then should return error",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusPending, mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()
				mockDisbursementRepo.On("RollbackTransaction", mock.Anything).Return(nil).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrSomeErrorForUnitTest,
		},
		{
			name:           "when failde to commit transaction, then should return error",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusPending, mock.Anything, mock.Anything).Return(nil).Once()
				mockDisbursementRepo.On("CommitTransaction", mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()
				mockDisbursementRepo.On("RollbackTransaction", mock.Anything).Return(nil).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrSomeErrorForUnitTest,
		},
		{
			name:           "when no issue, then should not return error ",
			disbursementID: validDisbursementID,
			payload:        validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusPending, mock.Anything, mock.Anything).Return(nil).Once()
								mockDisbursementRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				mockRedis.On("Del", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(redis.NewIntResult(0, nil)).Once()
			},
			shouldError: false,
		},
		{
			name:           "when success to transaction , then should not return error",
			disbursementID: validDisbursementID,
			payload: disbursementModel.ChangeDisbursementTransactionStatusRequest{
				DisbursementIDS: []string{validDisbursementID},
				Status:          constant.StatusFailed,
				ReasonType:      util.ValueToPtr(constant.ReasonTypeBlockedByHarsya),
			},
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusFailed, mock.Anything, mock.Anything).Return(nil).Once()
								mockDisbursementRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				mockRedis.On("Del", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(redis.NewIntResult(0, nil)).Once()
			},
			shouldError: false,
		},
		{
			name:           "when the ledger processed by snapcore and failed, then should return error",
			disbursementID: validDisbursementID,
			payload: disbursementModel.ChangeDisbursementTransactionStatusRequest{
				DisbursementIDS: []string{validDisbursementID},
				Status:          constant.StatusFailed,
				ReasonType:      util.ValueToPtr(constant.ReasonTypeBlockedByHarsya),
			},
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:               validUUID,
					MerchantID:         validUUID,
					ReferenceID:        validDisbursementID,
					Status:             constant.StatusPending,
					ProcessorReference: constant.SnapCoreProcessor,
				}, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusFailed, mock.Anything, mock.Anything).Return(nil).Once()
				mockSnapCoreRepo.On("UpdateBankTransferStatus", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest).Once()
				mockDisbursementRepo.On("RollbackTransaction", mock.Anything).Return(nil).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrSomeErrorForUnitTest,
		},
		{
			name:           "when the ledger processed by snapcore and success, then should return error",
			disbursementID: validDisbursementID,
			payload: disbursementModel.ChangeDisbursementTransactionStatusRequest{
				DisbursementIDS: []string{validDisbursementID},
				Status:          constant.StatusFailed,
				ReasonType:      util.ValueToPtr(constant.ReasonTypeBlockedByHarsya),
			},
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:               validUUID,
					MerchantID:         validUUID,
					ReferenceID:        validDisbursementID,
					Status:             constant.StatusPending,
					ProcessorReference: constant.SnapCoreProcessor,
				}, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusFailed, mock.Anything, mock.Anything).Return(nil).Once()
				mockSnapCoreRepo.On("UpdateBankTransferStatus", mock.Anything, mock.Anything).Return(nil).Once()
								mockDisbursementRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				mockRedis.On("Del", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(redis.NewIntResult(0, nil)).Once()
			},
			shouldError: false,
		},
		{
			name:           "when reference number is provided and bank reference update succeeds, then should not return error",
			disbursementID: validDisbursementID,
			payload: disbursementModel.ChangeDisbursementTransactionStatusRequest{
				DisbursementIDS: []string{validDisbursementID},
				Status:          constant.StatusPending,
				ReferenceNumber: "REF-12345",
			},
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusPending, mock.Anything, mock.Anything).Return(nil).Once()
				mockDisbursementRepo.On("UpdateBankReferenceNo", mock.Anything, validDisbursementID, "REF-12345").Return(nil).Once()
				mockDisbursementRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				mockRedis.On("Del", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(redis.NewIntResult(0, nil)).Once()
			},
			shouldError: false,
		},
		{
			name:           "when reference number is provided and bank reference update fails, then should return error",
			disbursementID: validDisbursementID,
			payload: disbursementModel.ChangeDisbursementTransactionStatusRequest{
				DisbursementIDS: []string{validDisbursementID},
				Status:          constant.StatusPending,
				ReferenceNumber: "REF-12345",
			},
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusPending, mock.Anything, mock.Anything).Return(nil).Once()
				mockDisbursementRepo.On("UpdateBankReferenceNo", mock.Anything, validDisbursementID, "REF-12345").Return(constant.ErrSomeErrorForUnitTest).Once()
				mockDisbursementRepo.On("RollbackTransaction", mock.Anything).Return(nil).Once()
			},
			shouldError: true,
			wantErr:     constant.ErrSomeErrorForUnitTest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			err := service.ProcessChangeDisbursementTansactionStatus(ctx, tc.disbursementID, tc.payload)

			if tc.shouldError {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
func TestChangeDisbursementTransactionStatus(t *testing.T) {
	ctx := context.Background()
	mockDisbursementRepo := repoMocks.NewIDisbursementRepository(t)
	mockOrchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	mockDisbursementInternalSvc := serviceMocks.NewIDisbursementInternalService(t)
	mockSnapCoreRepo := repoMocks.NewISnapCoreRepository(t)
	mockTransferSvc := serviceMocks.NewITransferService(t)
	mockFeeSvc := serviceMocks.NewIFeeService(t)
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	mockRedis := redisMocks.NewIRedisExt(t)
	mockStatusHistoriesRepo := repoMocks.NewIStatusHistoriesRepository(t)

	service := DisbursementService{
		logger:              mockLogger,
		disbursementRepo:    mockDisbursementRepo,
		orchestratorSvc:     mockOrchestratorSvc,
		self:                mockDisbursementInternalSvc,
		transferSvc:         mockTransferSvc,
		snapCoreRepo:        mockSnapCoreRepo,
		totalWorkerPool:     3,
		batchProcessWP:      nil,
		redisExt:            mockRedis,
		statusHistoriesRepo: mockStatusHistoriesRepo,
		feeSvc:              mockFeeSvc,
	}

	// General status history mock that will handle any calls
	mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

	wp, _ := ants.NewPoolWithFunc(3, service.batchProcessWPFunc)
	service.batchProcessWP = wp

	validUUID, _ := uuid.Parse(constant.EmptyUUID)
	feeID := uuid.New()
	validDisbursementID := "valid-disbursement-id"
	validRequest := disbursementModel.ChangeDisbursementTransactionStatusRequest{
		DisbursementIDS: []string{validDisbursementID},
		Status:          constant.StatusSuccess,
	}

	disbursementTransaction := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:       validDisbursementID,
			MerchantID: constant.EmptyUUID,
			Amount:     decimal.New(1000, 0),
		},
		TransactionStatus:     util.ValueToPtr(constant.StatusPending),
		TransactionReasonType: util.ValueToPtr(constant.ReasonTypeBlockedByHarsya),
	}

	disbursementLedger := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:               validUUID,
		MerchantID:         validUUID,
		ReferenceID:        validDisbursementID,
		Status:             constant.StatusPending,
		ProcessorReference: constant.SnapCoreProcessor,
	}

	testCases := []struct {
		name        string
		request     disbursementModel.ChangeDisbursementTransactionStatusRequest
		setupMock   func()
		expectedRes []disbursementModel.ChangeDisbursementTransactionStatusResponse
	}{
		{
			name:    "when failed to process change disbursement transaction status, then should return error response",
			request: validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedRes: []disbursementModel.ChangeDisbursementTransactionStatusResponse{
				{
					DisbursementID: validDisbursementID,
					Updated:        false,
					Reason:         constant.ErrSomeErrorForUnitTest.Error(),
				},
			},
		},
		{
			name:    "when successfully processed change disbursement transaction status, then should return success response",
			request: validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(nil, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()
				mockSnapCoreRepo.On("UpdateBankTransferStatus", mock.Anything, mock.Anything).Return(nil).Once()
								mockDisbursementRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				mockRedis.On("Del", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(redis.NewIntResult(0, nil)).Once()

			},
			expectedRes: []disbursementModel.ChangeDisbursementTransactionStatusResponse{
				{
					DisbursementID: validDisbursementID,
					Updated:        true,
					Reason:         "ok",
				},
			},
		},
		{
			name:    "when successfully processed change disbursement transaction status with non nil fee id, then should return success response",
			request: validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        feeID,
					MerchantID:  validUUID,
					ReferenceID: validDisbursementID,
					AdditionalInfo: types.NullJSONText{
						JSONText: []byte(`{transferId: "transfer-id"}`),
						Valid:    true,
					},
				}, nil).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, feeID.String(), constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()
				mockSnapCoreRepo.On("UpdateBankTransferStatus", mock.Anything, mock.Anything).Return(nil).Once()
								mockDisbursementRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				mockRedis.On("Del", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(redis.NewIntResult(0, nil)).Once()
			},
			expectedRes: []disbursementModel.ChangeDisbursementTransactionStatusResponse{
				{
					DisbursementID: validDisbursementID,
					Updated:        true,
					Reason:         "ok",
				},
			},
		},
		{
			name:    "when successfully processed with LADDER tiering counter, then should increment counter",
			request: validRequest,
			setupMock: func() {
				mockDisbursementRepo.On("FindByID", constant.ValueCtxMockType(), validDisbursementID).Return(disbursementTransaction, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeDisbursement).Return(disbursementLedger, nil).Once()
				mockOrchestratorSvc.On("FindByReference", constant.ValueCtxMockType(), validDisbursementID, constant.TypeFee).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:        feeID,
					MerchantID:  validUUID,
					ReferenceID: validDisbursementID,
					AdditionalInfo: types.NullJSONText{
						JSONText: []byte(`{"ladderCounterKey":"backend-portal:merchant-fee-counter:fee-uuid:2026-03","ladderCounterIncrement":500000}`),
						Valid:    true,
					},
				}, nil).Once()
				mockFeeSvc.On("IncrementLadderCounter", mock.Anything, "backend-portal:merchant-fee-counter:fee-uuid:2026-03", int64(500_000)).Once()
				mockDisbursementRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctx, nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, disbursementLedger.UUID.String(), constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()
				mockOrchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, feeID.String(), constant.StatusSuccess, mock.Anything, mock.Anything).Return(nil).Once()
				mockSnapCoreRepo.On("UpdateBankTransferStatus", mock.Anything, mock.Anything).Return(nil).Once()
				mockDisbursementRepo.On("CommitTransaction", mock.Anything).Return(nil).Once()
				mockRedis.On("Del", mock.Anything, constant.StringMockType(), constant.StringMockType()).Return(redis.NewIntResult(0, nil)).Once()
			},
			expectedRes: []disbursementModel.ChangeDisbursementTransactionStatusResponse{
				{
					DisbursementID: validDisbursementID,
					Updated:        true,
					Reason:         "ok",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			res := service.ChangeDisbursementTransactionStatus(ctx, tc.request)
			assert.Equal(t, tc.expectedRes, res)
		})
	}
}
