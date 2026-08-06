package unifiedPaymentService_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	rabbitMQMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResendPaymentCallback(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	merchantID := "merchant-" + uuid.NewString()
	paymentID := "payment-" + uuid.NewString()
	clientReferenceID := "client-ref-" + uuid.NewString()
	databaseError := errors.New("database error")

	validPayment := &paymentModel.Payment{
		UUID:       paymentID,
		MerchantID: merchantID,
		Amount:     decimal.NewFromFloat(100000),
		Currency:   "IDR",
		Status:     constant.UnifiedPaymentSessionStatusRequireAction,
		PaymentMethod: paymentModel.PaymentMethod{
			Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
		},
		Metadata: &map[string]interface{}{
			"paymentMethod": map[string]interface{}{
				"type": constant.UnifiedPaymentMethodCard,
			},
			"paymentMethodOptions": map[string]interface{}{
				"card": map[string]interface{}{
					"captureMethod": constant.UnifiedPaymentCardCaptureMethodManual,
				},
			},
			"isUnifiedPaymentV2": true,
		},
	}

	tests := []struct {
		name          string
		request       *callbackModel.ResendCallbackRequest
		setupMocks    func(*repositoryMock.IPaymentRepository, *repositoryMock.IAccountTransactionRepository, *rabbitMQMock.RabbitMQExt)
		expectedError error
	}{
		{
			name: "SUCCESS: Resend callback with ReferenceID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: paymentID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(validPayment, nil)
				accTrxRepo.On("FindByReference", mock.Anything, paymentID, constant.TypePayment).Return(&orchestratorModel.AccountTransactionWithUseCase{}, nil)
				// Callback mocks
				accTrxRepo.On("FindByReference", mock.Anything, paymentID, constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				rmq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rmq.On("PushNotification", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "SUCCESS: Resend callback with ClientReferenceID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:        merchantID,
				ClientReferenceID: clientReferenceID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, clientReferenceID).Return(validPayment, nil)
				// Callback mocks
				accTrxRepo.On("FindByReference", mock.Anything, paymentID, constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				rmq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rmq.On("PushNotification", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "ERROR: Database error when getting payment by ID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: paymentID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(nil, databaseError)
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, databaseError),
		},
		{
			name: "ERROR: Database error when getting payment by ClientReferenceID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:        merchantID,
				ClientReferenceID: clientReferenceID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, clientReferenceID).Return(nil, databaseError)
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, databaseError),
		},
		{
			name: "ERROR: Payment not found by ID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: paymentID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(nil, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrNotFound, constant.ErrPaymentNotFound),
		},
		{
			name: "ERROR: Payment not found by ClientReferenceID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:        merchantID,
				ClientReferenceID: clientReferenceID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, clientReferenceID).Return(nil, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrNotFound, constant.ErrPaymentNotFound),
		},
		{
			name: "ERROR: Merchant ID does not match",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: paymentID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: "different-merchant-" + uuid.NewString(),
					Status:     constant.UnifiedPaymentSessionStatusPaid,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrRequest, constant.ErrMerchantIsNotMatch),
		},
		{
			name: "ERROR: Payment is not unified payment v2 (nil metadata)",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: paymentID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     constant.UnifiedPaymentSessionStatusPaid,
					Metadata:   nil,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrRequest, errors.New("payment is not using unified payment v2")),
		},
		{
			name: "ERROR: Payment is not unified payment v2 (flag is false)",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: paymentID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				metadata := map[string]interface{}{
					"isUnifiedPaymentV2": false,
				}
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     constant.UnifiedPaymentSessionStatusPaid,
					Metadata:   &metadata,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrRequest, errors.New("payment is not using unified payment v2")),
		},
		{
			name: "ERROR: Payment is not unified payment v2 (metadata exists but no flag)",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: paymentID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				metadata := map[string]interface{}{
					"otherField": "someValue",
				}
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     constant.UnifiedPaymentSessionStatusPaid,
					Metadata:   &metadata,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrRequest, errors.New("payment is not using unified payment v2")),
		},
		{
			name: "SUCCESS: Resend callback with empty metadata",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: paymentID,
			},
			setupMocks: func(paymentRepo *repositoryMock.IPaymentRepository, accTrxRepo *repositoryMock.IAccountTransactionRepository, rmq *rabbitMQMock.RabbitMQExt) {
				emptyMetadata := map[string]interface{}{}
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     constant.UnifiedPaymentSessionStatusPaid,
					Metadata:   &emptyMetadata,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrRequest, errors.New("payment is not using unified payment v2")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			rabbitMq := rabbitMQMock.NewRabbitMQExt(t)

			// Setup mocks
			tt.setupMocks(paymentRepo, accountTrxRepo, rabbitMq)

			// Create service
			svc := unifiedPaymentService.New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo, unifiedPaymentService.WithRabbitMQClient(rabbitMq))

			// Execute
			err := svc.ResendPaymentCallback(context.Background(), tt.request)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			paymentRepo.AssertExpectations(t)
		})
	}
}
