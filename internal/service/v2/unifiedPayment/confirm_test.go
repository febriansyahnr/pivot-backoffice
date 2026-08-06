package unifiedPaymentService_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	snapCoreVaModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirmSession(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
		PaymentUIConfig: config.PaymentUIConfig{
			PaymentLinkURL: "link.here",
		},
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	merchantID := "cbfef85d-b54f-45de-ac7d-c6ac87baff31"
	parentMerchantId := uuid.NewString()

	requestVA := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
		PaymentSessionID: "payment-session-id",
		MerchantID:       parentMerchantId,
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{
			Type: c.UnifiedPaymentMethodVA,
		},
		PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
			VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
				Channel: "PERMATA", // NOSONAR
			},
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(
			paymentRepo *repositoryMock.IPaymentRepository,
			paymentMethodSvc *serviceMock.IPaymentMethodService,
			merchantRepo *repositoryMock.IMerchantRepository,
			snapCoreRepo *repositoryMock.ISnapCoreRepository,
			orchestratorSvc *serviceMock.IOrchestratorService,
			accountTrxRepo *repositoryMock.IAccountTransactionRepository,
			paymentSvc *serviceMock.IPaymentService,
			merchantSvc *serviceMock.IMerchantService,
		)
		input *unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest
	}{
		{
			name:    "ERROR: Got error database on GetPaymentById",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input: requestVA,
		},
		{
			name:    "ERROR: Payment is not found",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)
			},
			input: requestVA,
		},
		{
			name:    "ERROR: Merchant is not match",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					MerchantID: "different-merchant-id",
				}, nil)
			},
			input: requestVA,
		},
		{
			name:    "ERROR: Payment already processed",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusProcessing,
					MerchantID: parentMerchantId,
				}, nil)
			},
			input: requestVA,
		},
		{
			name:    "ERROR: Empty payment method chosen",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequirePaymentMethod,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
					MerchantID: parentMerchantId,
				}, nil)
			},
			input: &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
				PaymentSessionID: "payment-session-id",
				MerchantID:       parentMerchantId,
			},
		},
		{
			name:    "ERROR: Got error database on find merchant",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequirePaymentMethod,
					MerchantID: parentMerchantId,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
				}, nil)

				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input: requestVA,
		},
		{
			name:    "ERROR: Merchant is not found",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequirePaymentMethod,
					MerchantID: parentMerchantId,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
				}, nil)

				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)
			},
			input: requestVA,
		},
		{
			name:    "SUCCESS: Virtual Account with FindByReference error",
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				// First call for payment validation
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequirePaymentMethod,
					MerchantID: parentMerchantId,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
				}, nil)

				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchantModel.Merchant{}, nil)
				merchantSvc.On("GetFDSConfig", c.ValueCtxMockType(), c.StringMockType()).Return(nil, nil)
				paymentMethodSvc.On(
					"GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID: uuid.NewString(),
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.WithValue(context.Background(), c.CtxTest, ""), nil)

				// Mock FindByReference to return error and nil accountTransactions
				accountTrxRepo.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, c.ErrSomeErrorForUnitTest)

				snapCoreRepo.On("CreateVirtualAccount",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(&snapCoreVaModel.CreateVirtualAccountResponseData{}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					mock.Anything,
					mock.Anything,
				).Return(nil).Maybe()

				paymentRepo.On("UpdatePaymentData",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.PaymentDTO"),
				).Return(nil)

				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID",
					c.ValueCtxMockType(),
					mock.AnythingOfType("orchestrator_model.UpdatePaymentTransactionRequest"),
					mock.Anything).Return(nil)

				paymentSvc.On("RecordPaymentStatusHistory",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return()

				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				// Mock RollbackTransaction in case of error (optional but good practice)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Return(nil).Maybe()

				// Second call for response building
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					MerchantID: parentMerchantId,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
				}, nil)
			},
			input: requestVA,
		},
		{
			name:    "SUCCESS: Virtual Account with FindByReference with existing transaction",
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				// First call for payment validation
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequirePaymentMethod,
					MerchantID: parentMerchantId,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
				}, nil)

				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchantModel.Merchant{}, nil)
				merchantSvc.On("GetFDSConfig", c.ValueCtxMockType(), c.StringMockType()).Return(nil, nil)
				paymentMethodSvc.On(
					"GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID: uuid.NewString(),
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.WithValue(context.Background(), c.CtxTest, ""), nil)

				// Mock FindByReference to return existing account transaction (will skip PostAccountTransaction)
				accountTrxRepo.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)

				snapCoreRepo.On("CreateVirtualAccount",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(&snapCoreVaModel.CreateVirtualAccountResponseData{
					ID: uuid.NewString(),
				}, nil)

				// PostAccountTransaction should NOT be called when existing transaction is found

				paymentRepo.On("UpdatePaymentData",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.PaymentDTO"),
				).Return(nil)

				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID",
					c.ValueCtxMockType(),
					mock.AnythingOfType("orchestrator_model.UpdatePaymentTransactionRequest"),
					mock.Anything).Return(nil)

				paymentSvc.On("RecordPaymentStatusHistory",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return()

				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				// Mock RollbackTransaction in case of error (optional but good practice)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Return(nil).Maybe()

				// Second call for response building
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					MerchantID: parentMerchantId,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
				}, nil)

				orchestratorSvc.On("UpdateTransaction", mock.Anything, mock.Anything).Return(nil).Once()
			},
			input: requestVA,
		},
		{
			name:    "SUCCESS: Virtual Account",
			wantErr: false,
			input: &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
				PaymentSessionID: "payment-session-id",
				MerchantID:       parentMerchantId,
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodVA,
				},
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
						Channel: "PERMATA", // NOSONAR
					},
				},
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				// First call for payment validation
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequirePaymentMethod,
					MerchantID: parentMerchantId,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
				}, nil)

				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchantModel.Merchant{}, nil)
				merchantSvc.On("GetFDSConfig", c.ValueCtxMockType(), c.StringMockType()).Return(nil, nil)
				paymentMethodSvc.On(
					"GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID: uuid.NewString(),
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{},
				}, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.WithValue(context.Background(), c.CtxTest, ""), nil)

				// Mock FindByReference to return nil (no existing account transaction)
				accountTrxRepo.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				snapCoreRepo.On("CreateVirtualAccount",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(&snapCoreVaModel.CreateVirtualAccountResponseData{}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					mock.Anything,
					mock.Anything,
				).Return(nil).Maybe()

				paymentRepo.On("UpdatePaymentData",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.PaymentDTO"),
				).Return(nil)

				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID",
					c.ValueCtxMockType(),
					mock.AnythingOfType("orchestrator_model.UpdatePaymentTransactionRequest"),
					mock.Anything).Return(nil)

				paymentSvc.On("RecordPaymentStatusHistory",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return()

				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				// Mock RollbackTransaction in case of error (optional but good practice)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Return(nil).Maybe()

				// Second call for response building
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					MerchantID: parentMerchantId,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
				}, nil)
			},
		},
		{
			name:    "SUCCESS: Virtual Terminal",
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, paymentMethodSvc *serviceMock.IPaymentMethodService, merchantRepo *repositoryMock.IMerchantRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, orchestratorSvc *serviceMock.IOrchestratorService, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, merchantSvc *serviceMock.IMerchantService) {
				accountTrxRepo.On("FindByReference", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, mock.Anything).Return(nil, nil)
				orchestratorSvc.On("PostAccountTransaction", mock.Anything, mock.Anything).Return(nil)
				paymentSvc.On("RecordPaymentStatusHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				paymentRepo.On("UpdatePaymentData", mock.Anything, mock.Anything).Return(nil)
				paymentRepo.On(
					"GetPaymentById", mock.Anything, mock.Anything,
				).Return(&paymentModel.Payment{
					Status:     c.UnifiedPaymentSessionStatusRequireConfirmation,
					MerchantID: merchantID,
					Metadata:   &map[string]any{},
					PaymentURL: "link.here?token=token",
				}, nil)

				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{UUID: merchantID}, nil)

				paymentMethodSvc.On(
					"GetActivePaymentMethodDetailForPaymentRequest", mock.Anything, mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID: uuid.NewString(),
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{},
							},
						},
					},
				}, nil)
				paymentRepo.On("BeginTransaction", mock.Anything).Return(context.WithValue(context.Background(), c.CtxTest, ""), nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)
			},
			input: &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodCard,
				},
				PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
					Card: &unifiedPaymentModel.PaymentMethodOptionCard{
						ThreeDsMethod: c.CardThreeDsMethodNever,
						ProcessingConfig: &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
							BankMerchantId: "TEST0001", // NOSONAR
						},
					},
				},
				Mode: c.UnifiedPaymentModeRedirect,
				VirtualTerminal: &unifiedPaymentModel.VirtualTerminal{
					BatchID: "2f01f8f4-54ff-4c3c-b25d-417450e0acd2",
				},
				MerchantID: merchantID,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh mocks for each test case to avoid interference
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodSvc := serviceMock.NewIPaymentMethodService(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			merchantRepo := repositoryMock.NewIMerchantRepository(t)
			snapCoreRepo := repositoryMock.NewISnapCoreRepository(t)
			orchestratorSvc := serviceMock.NewIOrchestratorService(t)
			paymentSvc := serviceMock.NewIPaymentService(t)
			merchantSvc := serviceMock.NewIMerchantService(t)

			tc.setupMock(paymentRepo, paymentMethodSvc, merchantRepo, snapCoreRepo, orchestratorSvc, accountTrxRepo, paymentSvc, merchantSvc)

			svc := New(cfg, log, paymentRepo, nil, accountTrxRepo, WithMerchantRepo(merchantRepo), WithSnapCoreRepo(snapCoreRepo), WithOrchestratorService(orchestratorSvc), WithPaymentMethodService(paymentMethodSvc), WithPaymentService(paymentSvc))
			WithMerchantService(svc, merchantSvc)
			_, err := svc.ConfirmSession(context.Background(), tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				if strings.Contains(tc.name, "Payment Method Not Found") {
					assert.Contains(t, err.Error(), "payment method is not found")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
