package xbPayoutService

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	rabbitMqExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateStatusFromProcessor(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
		SlackConfig: config.SlackConfig{
			XBPayoutStatusUpdateWebhookURL: "https://slack.webhook.url",
		},
	}
	log, _ := logger.NewZapLogger(logger.Config{})
	ctx := context.Background()
	disbursementUUID := uuid.New().String()

	baseDisbursement := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:     disbursementUUID,
			Amount:   decimal.NewFromFloat(100),
			Currency: "USD",
			MetadataObj: disbursementModel.Metadata{
				XbDetail: &xbModel.XbPayoutMetadata{
					BeneficiaryData: xbModel.BeneficiaryDataResponse{
						Name:        "Test Beneficiary",
						CountryName: "Singapore",
					},
				},
				FeeDetail: feeModel.FeeMetadataObject{
					FinalAmount: 10.0,
				},
			},
		},
	}

	tests := []struct {
		name                string
		incomingStatus      string
		currentTrxStatus    string
		expectError         bool
		expectErrorContains string
		expectAllowsUpdate  bool
	}{
		{
			name:               "SUCCESS: RETURNED status updates final transaction (from FAILED)",
			incomingStatus:     constant.XbStatusReturned,
			currentTrxStatus:   constant.StatusFailed,
			expectError:        false,
			expectAllowsUpdate: true,
		},
		{
			name:               "SUCCESS: RETURNED status updates final transaction (from SUCCESS)",
			incomingStatus:     constant.XbStatusReturned,
			currentTrxStatus:   constant.StatusSuccess,
			expectError:        false,
			expectAllowsUpdate: true,
		},
		{
			name:                "ERROR: PAID status blocked on final transaction",
			incomingStatus:      constant.XbStatusPaid,
			currentTrxStatus:    constant.StatusSuccess,
			expectError:         true,
			expectErrorContains: constant.ErrTransactionAlreadyInFinalStatus.Error(),
			expectAllowsUpdate:  false,
		},
		{
			name:                "ERROR: ERROR status blocked on final transaction",
			incomingStatus:      constant.XbStatusError,
			currentTrxStatus:    constant.StatusFailed,
			expectError:         true,
			expectErrorContains: constant.ErrTransactionAlreadyInFinalStatus.Error(),
			expectAllowsUpdate:  false,
		},
		{
			name:               "SUCCESS: Normal status update on non-final transaction",
			incomingStatus:     constant.XbStatusPaid,
			currentTrxStatus:   constant.StatusPending,
			expectError:        false,
			expectAllowsUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
			beneficiaryRepo := repositoryMocks.NewIBeneficiaryAccountRepository(t)
			xbCoreRepo := repositoryMocks.NewIXbCoreProcessorRepository(t)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

			service := New(
				log,
				disbursementRepo,
				beneficiaryRepo,
				xbCoreRepo,
				WithOrchestratorService(orchestratorSvc),
				WithRabbitMQClient(rabbitMqExt),
				WithConfig(cfg),
				WithStatusHistories(statusHistoriesRepo),
			)

			request := &xbModel.ConsumePayoutStatusChangeRequest{
				AcquirerTransactionId: disbursementUUID,
				PartnerTransactionId:  "PARTNER-123",
				Status:                tt.incomingStatus,
				Timestamp:             time.Now(),
			}

			accountTransaction := &orchestratorModel.AccountTransactionWithUseCase{
				UUID:   uuid.New(),
				Status: tt.currentTrxStatus,
			}

			// Setup base mocks
			disbursementRepo.On("FindByProcessorReferenceID", mock.Anything, disbursementUUID).
				Return(baseDisbursement, nil)

			rabbitMqExt.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil).Maybe() // Slack notification

			orchestratorSvc.On("FindByReference", mock.Anything, disbursementUUID, constant.TypeDisbursement).
				Return(accountTransaction, nil)

			if tt.expectAllowsUpdate {
				orchestratorSvc.On("FindByReference", mock.Anything, disbursementUUID, constant.TypeFee).
					Return(accountTransaction, nil)

				// Setup mocks for successful update flow
				disbursementRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil)

				disbursementRepo.On("UpdateStatusAndReasonByID", mock.Anything, disbursementUUID, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				// For RETURNED after SUCCESS: use UpdateReasonOnly (no updated_at change)
				// For RETURNED after non-SUCCESS (e.g., FAILED): use UpdateStatusAccountTransaction
				// For other statuses when not SUCCESS: use UpdateStatusAccountTransaction
				isReturnedAfterSuccess := tt.incomingStatus == constant.XbStatusReturned && tt.currentTrxStatus == constant.StatusSuccess
				if isReturnedAfterSuccess {
					// Both disbursement and fee transactions will call UpdateReasonOnly
					orchestratorSvc.On("UpdateReasonOnly", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(2).Return(nil)
				} else if tt.currentTrxStatus != constant.StatusSuccess {
					orchestratorSvc.On("UpdateStatusAccountTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(2).Return(nil)
				}
				// Mock status history insert (called in recordStatusHistory)
				statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).
					Return(nil).Maybe()

				disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				disbursementRepo.On("FindByID", mock.Anything, disbursementUUID).
					Return(baseDisbursement, nil)

				rabbitMqExt.On("PublishMerchantCallback", mock.Anything, mock.Anything).
					Return(nil).Maybe() // Callback
			}

			// Execute
			err := service.UpdateStatusFromProcessor(ctx, request)

			// Verify
			if tt.expectError {
				require.Error(t, err)
				if tt.expectErrorContains != "" {
					assert.Contains(t, err.Error(), tt.expectErrorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateStatusFromProcessor_ReturnedAfterSuccessSkipsLedgerUpdate(t *testing.T) {
	// This test verifies the bug fix: when RETURNED notification comes after SUCCESS,
	// the ledger should NOT be updated at all (to preserve updated_at timestamp)
	// Operations will manually handle refund adjustments
	cfg := &config.Config{
		Environment: "test",
		SlackConfig: config.SlackConfig{
			XBPayoutStatusUpdateWebhookURL: "https://slack.webhook.url",
		},
	}
	log, _ := logger.NewZapLogger(logger.Config{})
	ctx := context.Background()
	disbursementUUID := uuid.New().String()

	baseDisbursement := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:     disbursementUUID,
			Amount:   decimal.NewFromFloat(100),
			Currency: "USD",
			MetadataObj: disbursementModel.Metadata{
				XbDetail: &xbModel.XbPayoutMetadata{
					BeneficiaryData: xbModel.BeneficiaryDataResponse{
						Name:        "Test Beneficiary",
						CountryName: "Singapore",
					},
				},
				FeeDetail: feeModel.FeeMetadataObject{
					FinalAmount: 10.0,
				},
			},
		},
	}

	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	beneficiaryRepo := repositoryMocks.NewIBeneficiaryAccountRepository(t)
	xbCoreRepo := repositoryMocks.NewIXbCoreProcessorRepository(t)
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

	service := New(
		log,
		disbursementRepo,
		beneficiaryRepo,
		xbCoreRepo,
		WithOrchestratorService(orchestratorSvc),
		WithRabbitMQClient(rabbitMqExt),
		WithConfig(cfg),
		WithStatusHistories(statusHistoriesRepo),
	)

	request := &xbModel.ConsumePayoutStatusChangeRequest{
		AcquirerTransactionId: disbursementUUID,
		PartnerTransactionId:  "PARTNER-123",
		Status:                constant.XbStatusReturned,
		Timestamp:             time.Now(),
	}

	// Account transaction with SUCCESS status (simulating after PAID notification)
	accountTransaction := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:   uuid.New(),
		Status: constant.StatusSuccess,
	}

	// Setup mocks
	disbursementRepo.On("FindByProcessorReferenceID", mock.Anything, disbursementUUID).
		Return(baseDisbursement, nil)

	rabbitMqExt.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe() // Slack notification

	orchestratorSvc.On("FindByReference", mock.Anything, disbursementUUID, constant.TypeDisbursement).
		Return(accountTransaction, nil)

	// Return nil for fee transaction (simulating no fee transaction)
	orchestratorSvc.On("FindByReference", mock.Anything, disbursementUUID, constant.TypeFee).
		Return(nil, nil)

	disbursementRepo.On("BeginTransaction", mock.Anything).Return(ctx, nil)

	disbursementRepo.On("UpdateStatusAndReasonByID", mock.Anything, disbursementUUID, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// IMPORTANT: UpdateReasonOnly should be called (updates reason_type without changing updated_at)
	// UpdateStatusAccountTransaction should NOT be called
	orchestratorSvc.On("UpdateReasonOnly", mock.Anything, accountTransaction.UUID.String(), mock.Anything, mock.Anything).
		Return(nil).Once()

	statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).
		Return(nil).Maybe()

	disbursementRepo.On("CommitTransaction", mock.Anything).Return(nil)

	disbursementRepo.On("FindByID", mock.Anything, disbursementUUID).
		Return(baseDisbursement, nil)

	rabbitMqExt.On("PublishMerchantCallback", mock.Anything, mock.Anything).
		Return(nil).Maybe()

	// Execute
	err := service.UpdateStatusFromProcessor(ctx, request)

	// Verify
	require.NoError(t, err)

	// Assert UpdateReasonOnly was called once
	orchestratorSvc.AssertCalled(t, "UpdateReasonOnly", mock.Anything, accountTransaction.UUID.String(), mock.Anything, mock.Anything)
	// Assert UpdateStatusAccountTransaction was NOT called
	orchestratorSvc.AssertNotCalled(t, "UpdateStatusAccountTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestSendCallbackToClient(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := logger.NewZapLogger(logger.Config{})
	ctx := context.Background()
	disbursementUUID := uuid.New().String()
	merchantID := "merchant-123"

	openAPICreatedFrom := constant.DisbursementCreatedFromOpenApi
	otherCreatedFrom := constant.DisbursementCreatedFromMerchantPortal

	tests := []struct {
		name                       string
		disbursement               *disbursementModel.DisbursementWithTransaction
		shouldPublishCallback      bool
		disbursementRepoFindError  error
		disbursementRepoFindResult *disbursementModel.DisbursementWithTransaction
	}{
		{
			name: "SUCCESS: Send callback for OpenAPI disbursement",
			disbursement: &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					UUID:        disbursementUUID,
					MerchantID:  merchantID,
					ReferenceID: "REF-123",
					Amount:      decimal.NewFromFloat(100),
					Currency:    "SGD",
					CreatedFrom: &openAPICreatedFrom,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					MetadataObj: disbursementModel.Metadata{
						XbDetail: &xbModel.XbPayoutMetadata{
							SourceCurrency:      "IDR",
							DestinationCurrency: "SGD",
							SourceAmount:        decimal.NewFromFloat(1500000),
							TotalAmount:         decimal.NewFromFloat(1510000),
							FxRate:              decimal.NewFromFloat(15000),
							DestinationFxRate:   decimal.NewFromFloat(1),
							ExpiredAt:           time.Now().Add(24 * time.Hour),
							PurposeCode:         "SALARY",
							BeneficiaryId:       "BEN-123",
							RoutingCode:         "BANK_CODE",
							RoutingValue:        "DBS",
							SenderData: xbModel.SenderDataResponse{
								Name:        "John Doe",
								Address:     "123 Street",
								City:        "Jakarta",
								CountryName: "Indonesia",
								CountryCode: "ID",
							},
							BeneficiaryData: xbModel.BeneficiaryDataResponse{
								Name:        "Jane Smith",
								CountryName: "Singapore",
							},
						},
						FeeDetail: feeModel.FeeMetadataObject{
							FinalAmount: 10.0,
						},
					},
				},
			},
			shouldPublishCallback: true,
		},
		{
			name: "SKIP: Don't send callback for non-OpenAPI disbursement (Merchant Portal)",
			disbursement: &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					UUID:        disbursementUUID,
					MerchantID:  merchantID,
					ReferenceID: "REF-456",
					Amount:      decimal.NewFromFloat(200),
					Currency:    "SGD",
					CreatedFrom: &otherCreatedFrom,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					MetadataObj: disbursementModel.Metadata{
						XbDetail: &xbModel.XbPayoutMetadata{
							SourceCurrency:      "IDR",
							DestinationCurrency: "SGD",
							SourceAmount:        decimal.NewFromFloat(3000000),
							TotalAmount:         decimal.NewFromFloat(3020000),
							FxRate:              decimal.NewFromFloat(15000),
							DestinationFxRate:   decimal.NewFromFloat(1),
							ExpiredAt:           time.Now().Add(24 * time.Hour),
							SenderData: xbModel.SenderDataResponse{
								Name: "John Doe",
							},
							BeneficiaryData: xbModel.BeneficiaryDataResponse{
								Name: "Jane Smith",
							},
						},
						FeeDetail: feeModel.FeeMetadataObject{
							FinalAmount: 20.0,
						},
					},
				},
			},
			shouldPublishCallback: false,
		},
		{
			name: "SKIP: Don't send callback when CreatedFrom is nil",
			disbursement: &disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					UUID:        disbursementUUID,
					MerchantID:  merchantID,
					ReferenceID: "REF-789",
					Amount:      decimal.NewFromFloat(300),
					Currency:    "SGD",
					CreatedFrom: nil,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					MetadataObj: disbursementModel.Metadata{
						XbDetail: &xbModel.XbPayoutMetadata{
							SourceCurrency:      "IDR",
							DestinationCurrency: "SGD",
							SourceAmount:        decimal.NewFromFloat(4500000),
							TotalAmount:         decimal.NewFromFloat(4530000),
							FxRate:              decimal.NewFromFloat(15000),
							DestinationFxRate:   decimal.NewFromFloat(1),
							ExpiredAt:           time.Now().Add(24 * time.Hour),
							SenderData: xbModel.SenderDataResponse{
								Name: "John Doe",
							},
							BeneficiaryData: xbModel.BeneficiaryDataResponse{
								Name: "Jane Smith",
							},
						},
						FeeDetail: feeModel.FeeMetadataObject{
							FinalAmount: 30.0,
						},
					},
				},
			},
			shouldPublishCallback: false,
		},
		{
			name:                       "ERROR: Disbursement not found by ID",
			disbursement:               nil,
			shouldPublishCallback:      false,
			disbursementRepoFindError:  nil,
			disbursementRepoFindResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
			beneficiaryRepo := repositoryMocks.NewIBeneficiaryAccountRepository(t)
			xbCoreRepo := repositoryMocks.NewIXbCoreProcessorRepository(t)
			rabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)

			service := xbPayoutService{
				logger:                 log,
				disbursementRepo:       disbursementRepo,
				beneficiaryAccountRepo: beneficiaryRepo,
				xbCoreProcessorRepo:    xbCoreRepo,
				rabbitMqExt:            rabbitMqExt,
				config:                 cfg,
			}

			// Setup mocks
			if tt.disbursementRepoFindResult != nil || tt.disbursementRepoFindError != nil {
				disbursementRepo.On("FindByID", mock.Anything, disbursementUUID).
					Return(tt.disbursementRepoFindResult, tt.disbursementRepoFindError)
			} else {
				disbursementRepo.On("FindByID", mock.Anything, disbursementUUID).
					Return(tt.disbursement, nil)
			}

			if tt.shouldPublishCallback {
				rabbitMqExt.On("PublishMerchantCallback", mock.Anything, mock.Anything).
					Return(nil).Once()
			}

			// Execute - Use reflection to call private method or test through public method
			// Since sendCallbackToClient is private, we need to call it through the service
			// We'll use the exported method by creating a wrapper or testing through UpdateStatusFromProcessor
			// For this test, we'll verify the behavior indirectly
			service.sendCallbackToClient(ctx, disbursementUUID)

			// Verify
			if tt.shouldPublishCallback {
				rabbitMqExt.AssertCalled(t, "PublishMerchantCallback", mock.Anything, mock.Anything)
			} else {
				rabbitMqExt.AssertNotCalled(t, "PublishMerchantCallback", mock.Anything, mock.Anything)
			}
		})
	}
}
