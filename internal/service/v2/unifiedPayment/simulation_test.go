package unifiedPaymentService_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSimulatePayment(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	merchantID := uuid.NewString()
	paymentSessionID := uuid.NewString()
	chargeID := uuid.NewString()

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(
			paymentRepo *repositoryMock.IPaymentRepository,
			accountTrxRepo *repositoryMock.IAccountTransactionRepository,
			paymentSvc *serviceMock.IPaymentService,
		)
		config *config.Config
		input  *unifiedPaymentModel.SimulatePaymentRequest
	}{
		{
			name:    "ERROR: Production environment",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				// No mocks needed, should fail before any repo calls
			},
			config: &config.Config{
				Environment: c.EnvironmentProduction,
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeStatus:     c.ChargeStatusSuccess,
			},
		},
		{
			name:    "ERROR: Database error on GetPaymentById",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeStatus:     c.ChargeStatusSuccess,
			},
		},
		{
			name:    "ERROR: Payment not found",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeStatus:     c.ChargeStatusSuccess,
			},
		},
		{
			name:    "ERROR: Merchant ID mismatch",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					MerchantID: "different-merchant-id",
				}, nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeStatus:     c.ChargeStatusSuccess,
			},
		},
		{
			name:    "ERROR: Invalid payment status",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusProcessing,
				}, nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeStatus:     c.ChargeStatusSuccess,
			},
		},
		{
			name:    "ERROR: Database error on FindByID",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
					Currency: "IDR",
				}, nil)

				accountTrxRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeID:         chargeID,
				ChargeStatus:     c.ChargeStatusSuccess,
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000,
				},
			},
		},
		{
			name:    "ERROR: Charge not found for REQUIRE_ACTION status",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
					Currency: "IDR",
				}, nil)

				accountTrxRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeID:         chargeID,
				ChargeStatus:     c.ChargeStatusSuccess,
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000,
				},
			},
		},
		{
			name:    "ERROR: Charge reference mismatch",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
					Currency: "IDR",
				}, nil)

				accountTrxRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: "different-payment-session-id",
				}, nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeID:         chargeID,
				ChargeStatus:     c.ChargeStatusSuccess,
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000,
				},
			},
		},
		{
			name:    "ERROR: Invalid payment method type",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
					Currency: "IDR",
				}, nil)

				accountTrxRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: paymentSessionID,
					Currency:    "IDR",
					Credit:      10000.0,
				}, nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeID:         chargeID,
				ChargeStatus:     c.ChargeStatusSuccess,
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000,
				},
			},
		},
		{
			name:    "ERROR: Failed status for non-ewallet payment method",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
					Currency: "IDR",
				}, nil)

				accountTrxRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: paymentSessionID,
					Currency:    "IDR",
					Credit:      10000.0,
				}, nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeID:         chargeID,
				ChargeStatus:     c.ChargeStatusFailed,
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000,
				},
			},
		},
		{
			name:    "SUCCESS: Simulate successful payment with REQUIRE_ACTION status",
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					UUID:       paymentSessionID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
					Currency: "IDR",
				}, nil)

				accountTrxRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: paymentSessionID,
					Currency:    "IDR",
					Credit:      10000.0,
				}, nil)

				paymentSvc.On("ProcessPaymentForSimulationByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
					mock.AnythingOfType("commonModel.Amount"),
					c.StringMockType(),
				).Once().Return(nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeID:         chargeID,
				ChargeStatus:     c.ChargeStatusSuccess,
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000,
				},
			},
		},
		{
			name:    "SUCCESS: Simulate successful payment with ACTIVE status (static payment)",
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					UUID:       paymentSessionID,
					MerchantID: merchantID,
					Status:     c.UnifiedStaticPaymentStatusActive,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
					Currency: "IDR",
				}, nil)

				accountTrxRepo.On("FindByReference",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)

				paymentSvc.On("ProcessPaymentForSimulationByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
					mock.AnythingOfType("commonModel.Amount"),
					c.StringMockType(),
				).Once().Return(nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeStatus:     c.ChargeStatusSuccess,
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000,
				},
			},
		},
		{
			name:    "SUCCESS: Simulate failed ewallet payment",
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService) {
				paymentRepo.On("GetPaymentById",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					UUID:       paymentSessionID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_EWALLET,
					},
					Currency: "IDR",
				}, nil)

				accountTrxRepo.On("FindByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					ReferenceID: paymentSessionID,
					Currency:    "IDR",
					Credit:      10000.0,
				}, nil)

				paymentSvc.On("ProcessPaymentForSimulationByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
					mock.AnythingOfType("commonModel.Amount"),
					c.StringMockType(),
				).Once().Return(nil)
			},
			config: &config.Config{
				Environment: "test",
			},
			input: &unifiedPaymentModel.SimulatePaymentRequest{
				PaymentSessionID: paymentSessionID,
				MerchantID:       merchantID,
				ChargeID:         chargeID,
				ChargeStatus:     c.ChargeStatusFailed,
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh mocks for each test case
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			paymentSvc := serviceMock.NewIPaymentService(t)

			tc.setupMock(paymentRepo, accountTrxRepo, paymentSvc)

			svc := New(
				tc.config,
				log,
				paymentRepo,
				nil,
				accountTrxRepo,
				WithPaymentService(paymentSvc),
				//WithRabbitMQClient(rabbitMq),
			)

			err := svc.SimulatePayment(context.Background(), tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				// For expired status, we can't fully test RabbitMQ without integration test
				// But we can verify no error is returned
				if tc.input.ChargeStatus == c.ChargeStatusExpired {
					// Allow error for RabbitMQ connection issues in unit test
					// In real scenario, this would be tested in integration test
					t.Log("Expired payment simulation - RabbitMQ publish may fail in unit test")
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}
