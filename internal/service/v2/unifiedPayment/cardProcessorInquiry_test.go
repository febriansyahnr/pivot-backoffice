package unifiedPaymentService_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	rabbitMqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInquiryCardPayment(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	merchantID := uuid.NewString()
	parentMerchantID := uuid.NewString()
	paymentID := uuid.NewString()
	// customerID := uuid.NewString() // Not used in tests
	referenceID := "REF123"
	processorRefID := uuid.NewString()
	authorizationID := "AUTH123"
	acquirerTrxID := "ACQ123"
	transactionID := uuid.NewString()

	metadataInterface := map[string]any{
		"isUnifiedPaymentV2": true,
		"PaymentMethodOptions": map[string]any{
			"card": map[string]any{},
		},
		"PaymentMethod": map[string]any{
			"type": c.ChannelCreditCard,
		},
	}

	metadataInterfaceOnBehalf := map[string]any{
		"isUnifiedPaymentV2": true,
		"PaymentMethodOptions": map[string]any{
			"card": map[string]any{},
		},
		"PaymentMethod": map[string]any{
			"type": c.ChannelCreditCard,
		},
		"onBehalf": map[string]any{
			"parentMerchantId": parentMerchantID,
		},
	}

	nonUnifiedMetadata := map[string]any{
		"isUnifiedPaymentV2": false,
	}

	testCases := []struct {
		name      string
		payment   *paymentModel.Payment
		wantErr   bool
		setupMock func(
			*repositoryMock.IPaymentRepository,
			*repositoryMock.IMerchantRepository,
			*serviceMock.IOrchestratorService,
			*serviceMock.ICreditCardService,
			*serviceMock.IPaymentService,
			*rabbitMqExtMock.RabbitMQExt,
		)
	}{
		{
			name: "ERROR: Non-unified payment returns forbidden",
			payment: &paymentModel.Payment{
				UUID:       paymentID,
				MerchantID: merchantID,
				Status:     c.UnifiedPaymentSessionStatusProcessing,
				Metadata:   &nonUnifiedMetadata,
			},
			wantErr: true,
			setupMock: func(
				paymentRepo *repositoryMock.IPaymentRepository,
				merchantRepo *repositoryMock.IMerchantRepository,
				orchestratorSvc *serviceMock.IOrchestratorService,
				creditcardSvc *serviceMock.ICreditCardService,
				paymentSvc *serviceMock.IPaymentService,
				rabbitMqMock *rabbitMqExtMock.RabbitMQExt,
			) {
				// No mocks needed as it returns early
			},
		},
		{
			name: "Error: Failed to get payment charge from orchestrator",
			payment: &paymentModel.Payment{
				UUID:       paymentID,
				MerchantID: merchantID,
				Status:     c.UnifiedPaymentSessionStatusProcessing,
				Metadata:   &metadataInterface,
			},
			wantErr: false, // Function returns payment with nil error on orchestrator failure
			setupMock: func(
				paymentRepo *repositoryMock.IPaymentRepository,
				merchantRepo *repositoryMock.IMerchantRepository,
				orchestratorSvc *serviceMock.IOrchestratorService,
				creditcardSvc *serviceMock.ICreditCardService,
				paymentSvc *serviceMock.IPaymentService,
				rabbitMqMock *rabbitMqExtMock.RabbitMQExt,
			) {
				orchestratorSvc.On("FindByReference",
					mock.Anything,
					paymentID,
					c.ReferencePayment,
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "Error: Payment charge not found returns nil",
			payment: &paymentModel.Payment{
				UUID:       paymentID,
				MerchantID: merchantID,
				Status:     c.UnifiedPaymentSessionStatusProcessing,
				Metadata:   &metadataInterface,
			},
			wantErr: false,
			setupMock: func(
				paymentRepo *repositoryMock.IPaymentRepository,
				merchantRepo *repositoryMock.IMerchantRepository,
				orchestratorSvc *serviceMock.IOrchestratorService,
				creditcardSvc *serviceMock.ICreditCardService,
				paymentSvc *serviceMock.IPaymentService,
				rabbitMqMock *rabbitMqExtMock.RabbitMQExt,
			) {
				orchestratorSvc.On("FindByReference",
					mock.Anything,
					paymentID,
					c.ReferencePayment,
				).Return(nil, nil)
			},
		},
		{
			name: "ERROR: Credit card inquiry transaction fails",
			payment: &paymentModel.Payment{
				UUID:        paymentID,
				MerchantID:  merchantID,
				Status:      c.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: &referenceID,
				Metadata:    &metadataInterface,
			},
			wantErr: true,
			setupMock: func(
				paymentRepo *repositoryMock.IPaymentRepository,
				merchantRepo *repositoryMock.IMerchantRepository,
				orchestratorSvc *serviceMock.IOrchestratorService,
				creditcardSvc *serviceMock.ICreditCardService,
				paymentSvc *serviceMock.IPaymentService,
				rabbitMqMock *rabbitMqExtMock.RabbitMQExt,
			) {
				charge := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 uuid.New(),
					ProcessorReferenceId: processorRefID,
				}
				orchestratorSvc.On("FindByReference",
					mock.Anything,
					paymentID,
					c.ReferencePayment,
				).Return(charge, nil)

				creditcardSvc.On("InquiryTransaction",
					mock.Anything,
					&creditcardModel.InquiryTransactionRequest{
						MerchantID:           merchantID,
						ClientReferenceID:    referenceID,
						ProcessorReferenceID: processorRefID,
					},
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "SUCCESS: Payment still processing",
			payment: &paymentModel.Payment{
				UUID:        paymentID,
				MerchantID:  merchantID,
				Status:      c.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: &referenceID,
				Metadata:    &metadataInterface,
			},
			wantErr: false,
			setupMock: func(
				paymentRepo *repositoryMock.IPaymentRepository,
				merchantRepo *repositoryMock.IMerchantRepository,
				orchestratorSvc *serviceMock.IOrchestratorService,
				creditcardSvc *serviceMock.ICreditCardService,
				paymentSvc *serviceMock.IPaymentService,
				rabbitMqMock *rabbitMqExtMock.RabbitMQExt,
			) {
				charge := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 uuid.New(),
					ProcessorReferenceId: processorRefID,
				}
				orchestratorSvc.On("FindByReference",
					mock.Anything,
					paymentID,
					c.ReferencePayment,
				).Return(charge, nil)

				creditcardSvc.On("InquiryTransaction",
					mock.Anything,
					&creditcardModel.InquiryTransactionRequest{
						MerchantID:           merchantID,
						ClientReferenceID:    referenceID,
						ProcessorReferenceID: processorRefID,
					},
				).Return(&creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: c.UnifiedPaymentSessionStatusProcessing,
				}, nil)
			},
		},
		{
			name: "SUCCESS: on behalf Payment still processing",
			payment: &paymentModel.Payment{
				UUID:        paymentID,
				MerchantID:  merchantID,
				Status:      c.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: &referenceID,
				Metadata:    &metadataInterfaceOnBehalf,
			},
			wantErr: false,
			setupMock: func(
				paymentRepo *repositoryMock.IPaymentRepository,
				merchantRepo *repositoryMock.IMerchantRepository,
				orchestratorSvc *serviceMock.IOrchestratorService,
				creditcardSvc *serviceMock.ICreditCardService,
				paymentSvc *serviceMock.IPaymentService,
				rabbitMqMock *rabbitMqExtMock.RabbitMQExt,
			) {
				charge := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 uuid.New(),
					ProcessorReferenceId: processorRefID,
				}
				orchestratorSvc.On("FindByReference",
					mock.Anything,
					paymentID,
					c.ReferencePayment,
				).Return(charge, nil)

				creditcardSvc.On("InquiryTransaction",
					mock.Anything,
					&creditcardModel.InquiryTransactionRequest{
						MerchantID:           parentMerchantID,
						ClientReferenceID:    referenceID,
						ProcessorReferenceID: processorRefID,
					},
				).Return(&creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: c.UnifiedPaymentSessionStatusProcessing,
				}, nil)
			},
		},
		{
			name: "SUCCESS: Card payment success with local channel",
			payment: &paymentModel.Payment{
				UUID:        paymentID,
				MerchantID:  merchantID,
				CustomerID:  "", // Set empty to avoid customer lookup in callback
				Status:      c.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: &referenceID,
				Metadata:    &metadataInterface,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: c.ChannelCreditCard,
				},
			},
			wantErr: false,
			setupMock: func(
				paymentRepo *repositoryMock.IPaymentRepository,
				merchantRepo *repositoryMock.IMerchantRepository,
				orchestratorSvc *serviceMock.IOrchestratorService,
				creditcardSvc *serviceMock.ICreditCardService,
				paymentSvc *serviceMock.IPaymentService,
				rabbitMqMock *rabbitMqExtMock.RabbitMQExt,
			) {
				methodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						SaveForFutureUse: util.ValueToPtr(false), // Set to false to avoid storeFutureUse path
						AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
							AuthorizationID: authorizationID,
						},
					},
				}
				methodDetailsBytes, _ := json.Marshal(map[string]any{
					"methodDetail": methodDetails,
				})
				chargeAdditionalInfo := types.NullJSONText{
					JSONText: methodDetailsBytes,
					Valid:    true,
				}

				charge := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 uuid.New(),
					ProcessorReferenceId: processorRefID,
					AdditionalInfo:       chargeAdditionalInfo,
				}
				orchestratorSvc.On("FindByReference",
					mock.Anything,
					paymentID,
					c.ReferencePayment,
				).Return(charge, nil)

				merchant := &merchantModel.Merchant{
					BusinessCountry: sql.NullString{
						String: "ID",
						Valid:  true,
					},
				}
				merchantRepo.On("FindMerchantByID",
					mock.Anything,
					merchantID,
				).Return(merchant, nil)

				inquiry := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus:         c.ChargeStatusSuccess,
					TransactionID:         uuid.MustParse(transactionID),
					AcquirerTransactionID: acquirerTrxID,
					Amount:                decimal.NewFromFloat(10000),
					Currency:              "IDR",
					Updated:               time.Now().UTC(),
					CardData: &creditcardModel.CardDataRequest{
						First8Digit:    "12345678",
						Last4Digit:     "1234",
						CardBrand:      "VISA",
						CardType:       "CREDIT",
						CardIssuing:    "BCA",
						CountryCode:    "ID",
						ExpiryMonth:    "12",
						ExpiryYear:     "2025",
						Fingerprint:    "abc123",
						SavedFutureUse: false,
						CardHolderName: "John Doe",
					},
					AuthenticationData: &creditcardModel.PaymentNotificationAuthenticationDataRequest{
						ThreeDsVer:           "2.0",
						AuthenticationResult: "Y",
						EciCode:              "05",
					},
					AuthorizationData: &creditcardModel.PaymentNotificationAuthorizationDataRequest{
						AcquirerTransactionID: acquirerTrxID,
						TransactionReference:  "TRX123",
						Stan:                  "123456",
						AvsResult:             "Y",
						CvvResult:             "M",
						AcquirerResponseCode:  "00",
						AuthorizationID:       authorizationID,
					},
					ResponseCode: &creditcardModel.PaymentNotificationResponseCode{
						GatewayCode:           "00",
						GatewayRecommendation: "APPROVE",
					},
					BankMerchantID:       "BANK123",
					AuthenticationMethod: "3DS",
				}
				creditcardSvc.On("InquiryTransaction",
					mock.Anything,
					&creditcardModel.InquiryTransactionRequest{
						MerchantID:           merchantID,
						ClientReferenceID:    referenceID,
						ProcessorReferenceID: processorRefID,
					},
				).Return(inquiry, nil)

				// Transaction mocks
				mockCtx := context.Background()
				paymentRepo.On("BeginTransaction", mock.Anything).Return(mockCtx, nil)
				paymentRepo.On("CommitTransaction", mockCtx).Return(nil)

				// Payment fee determination
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Return(nil)

				// Update payment status
				paymentRepo.On("UpdatePaymentStatus",
					mockCtx,
					paymentID,
					merchantID,
					c.UnifiedPaymentSessionStatusPaid,
					mock.Anything,
				).Return(nil)

				// Update pending ledger
				paymentSvc.On("UpdatePendingLedger", mockCtx, mock.Anything, mock.Anything).Return(nil)

				// Callback - note that sendCallback is called at the end
				orchestratorSvc.On("FindByReference", mockCtx, paymentID, c.ReferencePayment).Return(charge, nil).Maybe()
				rabbitMqMock.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Maybe()
				rabbitMqMock.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Maybe()
			},
		},
		{
			name: "ERROR: Begin transaction fails",
			payment: &paymentModel.Payment{
				UUID:        paymentID,
				MerchantID:  merchantID,
				Status:      c.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: &referenceID,
				Metadata:    &metadataInterface,
			},
			wantErr: true,
			setupMock: func(
				paymentRepo *repositoryMock.IPaymentRepository,
				merchantRepo *repositoryMock.IMerchantRepository,
				orchestratorSvc *serviceMock.IOrchestratorService,
				creditcardSvc *serviceMock.ICreditCardService,
				paymentSvc *serviceMock.IPaymentService,
				rabbitMqMock *rabbitMqExtMock.RabbitMQExt,
			) {
				charge := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 uuid.New(),
					ProcessorReferenceId: processorRefID,
				}
				orchestratorSvc.On("FindByReference",
					mock.Anything,
					paymentID,
					c.ReferencePayment,
				).Return(charge, nil)

				inquiry := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: c.ChargeStatusSuccess,
				}
				creditcardSvc.On("InquiryTransaction",
					mock.Anything,
					&creditcardModel.InquiryTransactionRequest{
						MerchantID:           merchantID,
						ClientReferenceID:    referenceID,
						ProcessorReferenceID: processorRefID,
					},
				).Return(inquiry, nil)

				// Transaction failure
				paymentRepo.On("BeginTransaction", mock.Anything).Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			merchantRepo := repositoryMock.NewIMerchantRepository(t)
			orchestratorSvc := serviceMock.NewIOrchestratorService(t)
			creditcardSvc := serviceMock.NewICreditCardService(t)
			paymentSvc := serviceMock.NewIPaymentService(t)
			rabbitMqMock := rabbitMqExtMock.NewRabbitMQExt(t)

			svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
				WithMerchantRepo(merchantRepo),
				WithOrchestratorService(orchestratorSvc),
				WithCreditCardService(creditcardSvc),
				WithPaymentService(paymentSvc),
				WithRabbitMQClient(rabbitMqMock),
			)

			// Add mock for sendCallback - it calls accountTrxRepo.FindByReference
			if tc.name == "SUCCESS: Card payment success with local channel" {
				charge := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 uuid.New(),
					ProcessorReferenceId: processorRefID,
				}
				accountTrxRepo.On("FindByReference", mock.Anything, paymentID, c.TypePayment).Return(charge, nil).Maybe()
			}

			tc.setupMock(paymentRepo, merchantRepo, orchestratorSvc, creditcardSvc, paymentSvc, rabbitMqMock)

			result, err := svc.InquiryCardPayment(context.Background(), tc.payment)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
