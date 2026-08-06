package unifiedPaymentService

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEvaluateSplitPaymentOutcome(t *testing.T) {
	log := logMock.NewILogger(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	internalSvc := serviceMocks.NewIInternalUnifiedPaymentService(t)
	cache := redisExtMocks.NewIRedisExt(t)
	mutexer := redisExtMocks.NewIMutexer(t)

	service := UnifiedPaymentService{
		logger:                    log,
		paymentRepo:               paymentRepo,
		internalUnifiedPaymentSvc: internalSvc,
		cache:                     cache,
	}

	parentPaymentID := "parent-payment-123"
	merchantID := "merchant-456"

	tests := []struct {
		name      string
		payment   *paymentModel.Payment
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR: failed to acquire lock",
			payment: &paymentModel.Payment{
				MerchantID:  merchantID,
				ReferenceID: util.ValueToPtr(parentPaymentID),
			},
			setupMock: func() {
				cache.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mutexer)
				mutexer.On("LockContext", mock.Anything).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "failed to lock auto split mutex ", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS: missing summary - returns nil",
			payment: &paymentModel.Payment{
				MerchantID:  merchantID,
				ReferenceID: util.ValueToPtr(parentPaymentID),
			},
			setupMock: func() {
				cache.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mutexer)
				mutexer.On("LockContext", mock.Anything).Once().Return(nil)
				mutexer.On("UnlockContext", mock.Anything).Once().Return(true, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, &paymentModel.GetAutoSplitPaymentSummaryRequest{
					ReferenceID:     parentPaymentID,
					MerchantID:      merchantID,
					MaxDateCreation: maxPaymentCreatedDays,
				}).Once().Return(nil, nil)
			},
			wantError: nil,
		},
		{
			name: "ERROR: failed to get auto split summary",
			payment: &paymentModel.Payment{
				MerchantID:  merchantID,
				ReferenceID: util.ValueToPtr(parentPaymentID),
			},
			setupMock: func() {
				cache.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mutexer)
				mutexer.On("LockContext", mock.Anything).Once().Return(nil)
				mutexer.On("UnlockContext", mock.Anything).Once().Return(true, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, &paymentModel.GetAutoSplitPaymentSummaryRequest{
					ReferenceID:     parentPaymentID,
					MerchantID:      merchantID,
					MaxDateCreation: maxPaymentCreatedDays,
				}).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "failed to get auto split summary", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS: payment still in progress - not all charges completed",
			payment: &paymentModel.Payment{
				MerchantID:  merchantID,
				ReferenceID: util.ValueToPtr(parentPaymentID),
			},
			setupMock: func() {
				cache.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mutexer)
				mutexer.On("LockContext", mock.Anything).Once().Return(nil)
				mutexer.On("UnlockContext", mock.Anything).Once().Return(true, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, &paymentModel.GetAutoSplitPaymentSummaryRequest{
					ReferenceID:     parentPaymentID,
					MerchantID:      merchantID,
					MaxDateCreation: maxPaymentCreatedDays,
				}).Once().Return(&paymentModel.AutoSplitPaymentSummary{
					NumberOfCharges:           3,
					NumberOfSuccessfulCharges: 1,
					NumberOfFailedCharges:     0,
				}, nil)
				log.On("Info", mock.Anything, "auto split payment still in progress", mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS: no charges yet - zero charges",
			payment: &paymentModel.Payment{
				MerchantID:  merchantID,
				ReferenceID: util.ValueToPtr(parentPaymentID),
			},
			setupMock: func() {
				cache.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mutexer)
				mutexer.On("LockContext", mock.Anything).Once().Return(nil)
				mutexer.On("UnlockContext", mock.Anything).Once().Return(true, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, &paymentModel.GetAutoSplitPaymentSummaryRequest{
					ReferenceID:     parentPaymentID,
					MerchantID:      merchantID,
					MaxDateCreation: maxPaymentCreatedDays,
				}).Once().Return(&paymentModel.AutoSplitPaymentSummary{
					NumberOfCharges: 0,
				}, nil)
				log.On("Info", mock.Anything, "auto split payment still in progress", mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS: all charges completed - triggers FinalizeSplitPayment",
			payment: &paymentModel.Payment{
				MerchantID:  merchantID,
				ReferenceID: util.ValueToPtr(parentPaymentID),
			},
			setupMock: func() {
				cache.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mutexer)
				mutexer.On("LockContext", mock.Anything).Once().Return(nil)
				mutexer.On("UnlockContext", mock.Anything).Once().Return(true, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, &paymentModel.GetAutoSplitPaymentSummaryRequest{
					ReferenceID:     parentPaymentID,
					MerchantID:      merchantID,
					MaxDateCreation: maxPaymentCreatedDays,
				}).Once().Return(&paymentModel.AutoSplitPaymentSummary{
					NumberOfCharges:           2,
					NumberOfSuccessfulCharges: 1,
					NumberOfFailedCharges:     1,
				}, nil)

				internalSvc.On("FinalizeSplitPayment", mock.Anything, &paymentModel.ProcessSplitPaymentRequest{
					ParentPaymentID: parentPaymentID,
					MerchantID:      merchantID,
					Summary: &paymentModel.AutoSplitPaymentSummary{
						NumberOfCharges:           2,
						NumberOfSuccessfulCharges: 1,
						NumberOfFailedCharges:     1,
					},
				}).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "ERROR: FinalizeSplitPayment returns error",
			payment: &paymentModel.Payment{
				MerchantID:  merchantID,
				ReferenceID: util.ValueToPtr(parentPaymentID),
			},
			setupMock: func() {
				cache.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mutexer)
				mutexer.On("LockContext", mock.Anything).Once().Return(nil)
				mutexer.On("UnlockContext", mock.Anything).Once().Return(true, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, &paymentModel.GetAutoSplitPaymentSummaryRequest{
					ReferenceID:     parentPaymentID,
					MerchantID:      merchantID,
					MaxDateCreation: maxPaymentCreatedDays,
				}).Once().Return(&paymentModel.AutoSplitPaymentSummary{
					NumberOfCharges:           3,
					NumberOfSuccessfulCharges: 3,
					NumberOfFailedCharges:     0,
				}, nil)

				internalSvc.On("FinalizeSplitPayment", mock.Anything, &paymentModel.ProcessSplitPaymentRequest{
					ParentPaymentID: parentPaymentID,
					MerchantID:      merchantID,
					Summary: &paymentModel.AutoSplitPaymentSummary{
						NumberOfCharges:           3,
						NumberOfSuccessfulCharges: 3,
						NumberOfFailedCharges:     0,
					},
				}).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			err := service.EvaluateSplitPaymentOutcome(t.Context(), tt.payment)

			require.Equal(t, tt.wantError, err)

			paymentRepo.AssertExpectations(t)
			log.AssertExpectations(t)
			internalSvc.AssertExpectations(t)
			cache.AssertExpectations(t)
			mutexer.AssertExpectations(t)
		})
	}
}

func TestGetAutoSplitPaymentDetail(t *testing.T) {
	log := logMock.NewILogger(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)

	service := UnifiedPaymentService{
		logger:      log,
		paymentRepo: paymentRepo,
	}

	referenceID := "ref-123"
	merchantID := "merchant-456"

	payment := &paymentModel.Payment{
		Status: constant.UnifiedPaymentSessionStatusPaid,
	}

	tests := []struct {
		name        string
		request     *paymentModel.GetAutoSplitPaymentSummaryRequest
		setupMock   func()
		wantError   error
		wantSummary *unifiedPaymentModel.AutoSplitPaymentSummary
	}{
		{
			name: "ERROR: failed to get parent payment ",
			request: &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxPaymentCreatedDays,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "failed to get parent payment", mock.Anything).Once().Return()
			},
			wantError:   assert.AnError,
			wantSummary: nil,
		},
		{
			name: "ERROR: parent payment not found",
			request: &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxPaymentCreatedDays,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Once().Return(nil, nil)
			},
			wantError:   constant.ErrDataNotFound,
			wantSummary: nil,
		},
		{
			name: "ERROR: failed to get auto split summary",
			request: &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxPaymentCreatedDays,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Once().Return(payment, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, mock.MatchedBy(func(req *paymentModel.GetAutoSplitPaymentSummaryRequest) bool {
					return req.ReferenceID == referenceID && req.MerchantID == merchantID && req.MaxDateCreation == maxPaymentCreatedDays
				})).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "failed to get auto split summary", mock.Anything).Once().Return()
			},
			wantError:   assert.AnError,
			wantSummary: nil,
		},
		{
			name: "ERROR: failed to get charges",
			request: &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxPaymentCreatedDays,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Once().Return(payment, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, mock.Anything).Once().Return(&paymentModel.AutoSplitPaymentSummary{
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   1,
					NumberOfFailedCharges:       1,
					TotalSuccessfulChargeAmount: 50000.0,
					TotalFailedChargeAmount:     25000.0,
					TotalInProgressChargeAmount: 0,
				}, nil)

				paymentRepo.On("GetCharges", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.FilterChargeRequest) bool {
					return req.MerchantID == merchantID && req.ClientReferenceID == referenceID
				})).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "failed to get charges", mock.Anything, mock.Anything).Once().Return()
			},
			wantError:   assert.AnError,
			wantSummary: nil,
		},
		{
			name: "SUCCESS: summary not found, should return nil",
			request: &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxPaymentCreatedDays,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Once().Return(payment, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantError:   nil,
			wantSummary: nil,
		},
		{
			name: "SUCCESS: all charges successful",
			request: &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxPaymentCreatedDays,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Once().Return(payment, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, mock.Anything).Once().Return(&paymentModel.AutoSplitPaymentSummary{
					NumberOfCharges:             3,
					NumberOfSuccessfulCharges:   3,
					NumberOfFailedCharges:       0,
					NumberOfInProcessCharges:    0,
					TotalSuccessfulChargeAmount: 150000.0,
					TotalFailedChargeAmount:     0,
					TotalInProgressChargeAmount: 0,
				}, nil)

				paymentRepo.On("GetCharges", mock.Anything, mock.Anything).Once().Return([]unifiedPaymentModel.ChargeResponse{
					{ID: "charge-1"},
					{ID: "charge-2"},
					{ID: "charge-3"},
				}, nil)
			},
			wantError: nil,
			wantSummary: &unifiedPaymentModel.AutoSplitPaymentSummary{
				Status:                      "SUCCESS",
				NumberOfCharges:             3,
				NumberOfSuccessfulCharges:   3,
				NumberOfFailedCharges:       0,
				NumberOfInProcessCharges:    0,
				TotalSuccessfulChargeAmount: commonModel.Amount{Currency: "IDR", Value: "150000.00"},
				TotalFailedChargeAmount:     commonModel.Amount{Currency: "IDR", Value: "0.00"},
				TotalInProgressChargeAmount: commonModel.Amount{Currency: "IDR", Value: "0.00"},
				ChargeDetails: []unifiedPaymentModel.ChargeResponse{
					{ID: "charge-1"},
					{ID: "charge-2"},
					{ID: "charge-3"},
				},
			},
		},
		{
			name: "SUCCESS: all charges failed",
			request: &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxPaymentCreatedDays,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Once().Return(&paymentModel.Payment{Status: constant.UnifiedPaymentSessionStatusCancelled}, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, mock.Anything).Once().Return(&paymentModel.AutoSplitPaymentSummary{
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   0,
					NumberOfFailedCharges:       2,
					NumberOfInProcessCharges:    0,
					TotalSuccessfulChargeAmount: 0,
					TotalFailedChargeAmount:     100000.0,
					TotalInProgressChargeAmount: 0,
				}, nil)

				paymentRepo.On("GetCharges", mock.Anything, mock.Anything).Once().Return([]unifiedPaymentModel.ChargeResponse{
					{ID: "charge-1"},
					{ID: "charge-2"},
				}, nil)
			},
			wantError: nil,
			wantSummary: &unifiedPaymentModel.AutoSplitPaymentSummary{
				Status:                      "FAILED",
				NumberOfCharges:             2,
				NumberOfSuccessfulCharges:   0,
				NumberOfFailedCharges:       2,
				NumberOfInProcessCharges:    0,
				TotalSuccessfulChargeAmount: commonModel.Amount{Currency: "IDR", Value: "0.00"},
				TotalFailedChargeAmount:     commonModel.Amount{Currency: "IDR", Value: "100000.00"},
				TotalInProgressChargeAmount: commonModel.Amount{Currency: "IDR", Value: "0.00"},
				ChargeDetails: []unifiedPaymentModel.ChargeResponse{
					{ID: "charge-1"},
					{ID: "charge-2"},
				},
			},
		},
		{
			name: "SUCCESS: partial success with mixed results",
			request: &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxPaymentCreatedDays,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Once().Return(payment, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, mock.Anything).Once().Return(&paymentModel.AutoSplitPaymentSummary{
					NumberOfCharges:             3,
					NumberOfSuccessfulCharges:   2,
					NumberOfFailedCharges:       1,
					NumberOfInProcessCharges:    0,
					TotalSuccessfulChargeAmount: 100000.0,
					TotalFailedChargeAmount:     50000.0,
					TotalInProgressChargeAmount: 0,
				}, nil)

				paymentRepo.On("GetCharges", mock.Anything, mock.Anything).Once().Return([]unifiedPaymentModel.ChargeResponse{
					{ID: "charge-1"},
					{ID: "charge-2"},
					{ID: "charge-3"},
				}, nil)
			},
			wantError: nil,
			wantSummary: &unifiedPaymentModel.AutoSplitPaymentSummary{
				Status:                      "PARTIAL_SUCCESS",
				NumberOfCharges:             3,
				NumberOfSuccessfulCharges:   2,
				NumberOfFailedCharges:       1,
				NumberOfInProcessCharges:    0,
				TotalSuccessfulChargeAmount: commonModel.Amount{Currency: "IDR", Value: "100000.00"},
				TotalFailedChargeAmount:     commonModel.Amount{Currency: "IDR", Value: "50000.00"},
				TotalInProgressChargeAmount: commonModel.Amount{Currency: "IDR", Value: "0.00"},
				ChargeDetails: []unifiedPaymentModel.ChargeResponse{
					{ID: "charge-1"},
					{ID: "charge-2"},
					{ID: "charge-3"},
				},
			},
		},
		{
			name: "SUCCESS: still processing - all in progress",
			request: &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxPaymentCreatedDays,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, referenceID).Once().Return(&paymentModel.Payment{Status: constant.UnifiedPaymentSessionStatusProcessing}, nil)
				paymentRepo.On("GetSummaryAutoSplitPayment", mock.Anything, mock.Anything).Once().Return(&paymentModel.AutoSplitPaymentSummary{
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   0,
					NumberOfFailedCharges:       0,
					NumberOfInProcessCharges:    2,
					TotalSuccessfulChargeAmount: 0,
					TotalFailedChargeAmount:     0,
					TotalInProgressChargeAmount: 100000.0,
				}, nil)

				paymentRepo.On("GetCharges", mock.Anything, mock.Anything).Once().Return([]unifiedPaymentModel.ChargeResponse{}, nil)
			},
			wantError: nil,
			wantSummary: &unifiedPaymentModel.AutoSplitPaymentSummary{
				Status:                      "PROCESSING",
				NumberOfCharges:             2,
				NumberOfSuccessfulCharges:   0,
				NumberOfFailedCharges:       0,
				NumberOfInProcessCharges:    2,
				TotalSuccessfulChargeAmount: commonModel.Amount{Currency: "IDR", Value: "0.00"},
				TotalFailedChargeAmount:     commonModel.Amount{Currency: "IDR", Value: "0.00"},
				TotalInProgressChargeAmount: commonModel.Amount{Currency: "IDR", Value: "100000.00"},
				ChargeDetails:               []unifiedPaymentModel.ChargeResponse{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			result, err := service.GetAutoSplitPaymentDetail(t.Context(), tt.request)

			require.Equal(t, tt.wantError, err)
			require.Equal(t, tt.wantSummary, result)

			paymentRepo.AssertExpectations(t)
			log.AssertExpectations(t)
		})
	}
}

func TestFinalizeSplitPayment(t *testing.T) {
	log := logMock.NewILogger(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)
	internalSvc := serviceMocks.NewIInternalUnifiedPaymentService(t)
	paymentSvc := serviceMocks.NewIPaymentService(t)
	rabbitMqExt := rabbitMqMocks.NewRabbitMQExt(t)

	svc := UnifiedPaymentService{
		logger:                    log,
		paymentRepo:               paymentRepo,
		accountTransactionRepo:    accountTransactionRepo,
		internalUnifiedPaymentSvc: internalSvc,
		paymentSvc:                paymentSvc,
		rabbitMqExt:               rabbitMqExt,
	}

	parentPaymentID := "parent-payment-123"
	merchantID := "merchant-456"

	paymentUUID := uuid.New()
	now := time.Now().UTC()
	trxDatetime := now.Add(-1 * time.Hour)

	autoSplitMetadata := unifiedPaymentModel.AutoSplitPayment{
		Summary: &unifiedPaymentModel.AutoSplitPaymentSummary{},
	}

	feeDetail := &feeModel.FeeMetadataObject{}

	basePayment := &paymentModel.Payment{
		UUID:                   paymentUUID.String(),
		MerchantID:             merchantID,
		ReferenceID:            util.ValueToPtr(parentPaymentID),
		Currency:               "IDR",
		Status:                 "PROCESSING",
		Amount:                 decimal.NewFromInt(150000),
		TotalAmount:            decimal.NewFromInt(150000),
		ProcessorID:            "proc-123",
		ProcessorTransactionID: "txn-123",
		TrxDatetime:            &trxDatetime,
		CreatedAt:              now.Add(-2 * time.Hour),
		UpdatedAt:              now.Add(-1 * time.Hour),
		Metadata:               &map[string]any{"autoSplitPayment": autoSplitMetadata, "feeDetail": feeDetail},
	}

	baseLedger := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:            paymentUUID,
		ReferenceID:     parentPaymentID,
		SettlementModel: sql.NullString{String: "NEXT_DAY", Valid: true},
	}

	ctx := context.Background()

	tests := []struct {
		name      string
		request   *paymentModel.ProcessSplitPaymentRequest
		setupMock func()
		wantError bool
	}{
		{
			name: "ERROR: failed to get parent payment",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to get parent payment for split", mock.Anything, mock.Anything).Once().Return()
			},
			wantError: true,
		},
		{
			name: "ERROR: parent payment not found",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(nil, nil)
				log.On("Error", mock.Anything, "Parent payment not found for split", mock.Anything).Once().Return()
			},
			wantError: true,
		},
		{
			name: "ERROR: failed to get auto split payment detail",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(util.ClonePtr(basePayment), nil)
				internalSvc.On("GetAutoSplitPaymentDetail", mock.Anything, mock.MatchedBy(func(req *paymentModel.GetAutoSplitPaymentSummaryRequest) bool {
					return req.ReferenceID == parentPaymentID && req.MerchantID == merchantID
				})).Once().Return(nil, assert.AnError)
			},
			wantError: true,
		},
		{
			name: "ERROR: failed to get ledger",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(util.ClonePtr(basePayment), nil)
				internalSvc.On("GetAutoSplitPaymentDetail", mock.Anything, mock.Anything).Once().Return(&unifiedPaymentModel.AutoSplitPaymentSummary{
					Status:                      "SUCCESS",
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   2,
					TotalSuccessfulChargeAmount: commonModel.Amount{},
				}, nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "failed to get ledger", mock.Anything).Once().Return()
			},
			wantError: true,
		},
		{
			name: "ERROR: ledger not found",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(util.ClonePtr(basePayment), nil)
				internalSvc.On("GetAutoSplitPaymentDetail", mock.Anything, mock.Anything).Once().Return(&unifiedPaymentModel.AutoSplitPaymentSummary{
					Status:                      "SUCCESS",
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   2,
					TotalSuccessfulChargeAmount: commonModel.Amount{},
				}, nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(nil, nil)
				log.On("Error", mock.Anything, "ledger not found", mock.Anything).Once().Return()
			},
			wantError: true,
		},
		{
			name: "ERROR: failed to begin transaction",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(util.ClonePtr(basePayment), nil)
				internalSvc.On("GetAutoSplitPaymentDetail", mock.Anything, mock.Anything).Once().Return(&unifiedPaymentModel.AutoSplitPaymentSummary{
					Status:                      "SUCCESS",
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   2,
					TotalSuccessfulChargeAmount: commonModel.Amount{},
				}, nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(baseLedger, nil)
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctx, assert.AnError)
			},
			wantError: true,
		},
		{
			name: "ERROR: failed to update payment data",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(util.ClonePtr(basePayment), nil)
				internalSvc.On("GetAutoSplitPaymentDetail", mock.Anything, mock.Anything).Once().Return(&unifiedPaymentModel.AutoSplitPaymentSummary{
					Status:                      "SUCCESS",
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   2,
					TotalSuccessfulChargeAmount: commonModel.Amount{},
				}, nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(baseLedger, nil)
				ctxTrx := context.Background()
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(nil)
				paymentRepo.On("UpdatePaymentData", ctxTrx, mock.Anything).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "failed to update payment", mock.Anything).Once().Return()
			},
			wantError: true,
		},
		{
			name: "ERROR: failed to update pending ledger",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(util.ClonePtr(basePayment), nil)
				internalSvc.On("GetAutoSplitPaymentDetail", mock.Anything, mock.Anything).Once().Return(&unifiedPaymentModel.AutoSplitPaymentSummary{
					Status:                      "SUCCESS",
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   2,
					TotalSuccessfulChargeAmount: commonModel.Amount{},
				}, nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(baseLedger, nil)
				ctxTrx := context.Background()
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(nil)
				paymentRepo.On("UpdatePaymentData", ctxTrx, mock.Anything).Once().Return(nil)
				paymentSvc.On("UpdatePendingLedger", ctxTrx, mock.Anything, mock.Anything).Once().Return(assert.AnError)
			},
			wantError: true,
		},
		{
			name: "ERROR: failed to commit transaction",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(util.ClonePtr(basePayment), nil)
				internalSvc.On("GetAutoSplitPaymentDetail", mock.Anything, mock.Anything).Once().Return(&unifiedPaymentModel.AutoSplitPaymentSummary{
					Status:                      "SUCCESS",
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   2,
					TotalSuccessfulChargeAmount: commonModel.Amount{},
				}, nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(baseLedger, nil)
				ctxTrx := context.Background()
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("UpdatePaymentData", ctxTrx, mock.Anything).Once().Return(nil)
				paymentSvc.On("UpdatePendingLedger", ctxTrx, mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("CommitTransaction", ctxTrx).Once().Return(assert.AnError)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(nil)
			},
			wantError: true,
		},
		{
			name: "SUCCESS: finalize split payment on paid payment status",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				payment := *util.ClonePtr(basePayment)
				payment.Status = constant.UnifiedPaymentSessionStatusPaid
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(&payment, nil)
				log.On("Warn", mock.Anything, "auto split payment already paid, skip the process").Once().Return()
			},
			wantError: false,
		},
		{
			name: "SUCCESS: finalize split payment with all charges successful",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(util.ClonePtr(&paymentModel.Payment{
					UUID:                   paymentUUID.String(),
					MerchantID:             merchantID,
					ReferenceID:            new(parentPaymentID),
					Currency:               "IDR",
					Status:                 "PROCESSING",
					Amount:                 decimal.NewFromInt(150000),
					TotalAmount:            decimal.NewFromInt(150000),
					ProcessorID:            "proc-123",
					ProcessorTransactionID: "txn-123",
					TrxDatetime:            &trxDatetime,
					CreatedAt:              now.Add(-2 * time.Hour),
					UpdatedAt:              now.Add(-1 * time.Hour),
					Metadata:               &map[string]any{"autoSplitPayment": autoSplitMetadata, "feeDetail": feeDetail},
				}), nil)
				internalSvc.On("GetAutoSplitPaymentDetail", mock.Anything, mock.Anything).Once().Return(&unifiedPaymentModel.AutoSplitPaymentSummary{
					Status:                      "SUCCESS",
					NumberOfCharges:             2,
					NumberOfSuccessfulCharges:   2,
					TotalSuccessfulChargeAmount: commonModel.Amount{},
				}, nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(baseLedger, nil)
				ctxTrx := context.Background()
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("UpdatePaymentData", ctxTrx, mock.Anything).Once().Return(nil)
				paymentSvc.On("UpdatePendingLedger", ctxTrx, mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("CommitTransaction", ctxTrx).Once().Return(nil)
				rabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Once().Return(nil)

				// break send callback flow
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(nil, nil)
			},
			wantError: false,
		},
		{
			name: "SUCCESS: finalize split payment with partial success - notification error logged",
			request: &paymentModel.ProcessSplitPaymentRequest{
				ParentPaymentID: parentPaymentID,
				MerchantID:      merchantID,
			},
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, parentPaymentID).Once().Return(util.ClonePtr(basePayment), nil)
				internalSvc.On("GetAutoSplitPaymentDetail", mock.Anything, mock.Anything).Once().Return(&unifiedPaymentModel.AutoSplitPaymentSummary{
					Status:                    "PARTIAL_SUCCESS",
					NumberOfCharges:           3,
					NumberOfSuccessfulCharges: 2,
					NumberOfFailedCharges:     1,
				}, nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(baseLedger, nil)
				ctxTrx := context.Background()
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("UpdatePaymentData", ctxTrx, mock.Anything).Once().Return(nil)
				paymentSvc.On("UpdatePendingLedger", ctxTrx, mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("CommitTransaction", ctxTrx).Once().Return(nil)
				rabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "failed to push notification for payment", mock.Anything, mock.Anything).Once().Return()

				// break send callback flow
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentUUID.String(), mock.Anything).
					Once().Return(nil, nil)
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			err := svc.FinalizeSplitPayment(t.Context(), tt.request)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			paymentRepo.AssertExpectations(t)
			log.AssertExpectations(t)
			internalSvc.AssertExpectations(t)
			paymentSvc.AssertExpectations(t)
			rabbitMqExt.AssertExpectations(t)
		})
	}
}
