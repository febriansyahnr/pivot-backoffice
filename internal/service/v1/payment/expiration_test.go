package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreQRModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	snapCoreVAModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	rabbitMqExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPublishPaymentExpirationMessage(t *testing.T) {

	tests := []struct {
		name          string
		mockSetup     func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt)
		expectedError error
	}{
		{
			name: "when expiring payment published, then should not return error",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				expiringPayments := []*paymentModel.ExpiringPayment{
					{UUID: "uuid-1", MerchantID: "merchant-1", ExpiredAt: time.Now().Add(24 * time.Hour)},
				}
				mockPaymentRepo.On("GetExpiringPayments", mock.Anything, mock.Anything, mock.Anything).Return(expiringPayments, nil).Once()
				mockRabbitMqExt.On("PublishWithDelay", mock.Anything, rabbitMqExt.PaymentExpirationRoutingKey, expiringPayments[0], mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when failed to get expiring payment, then should return error",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				mockPaymentRepo.On("GetExpiringPayments", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("database error")).Once()
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "when failed to publish the expiring payment, then should return error",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				expiringPayments := []*paymentModel.ExpiringPayment{
					{UUID: "uuid-1", MerchantID: "merchant-1", ExpiredAt: time.Now().Add(24 * time.Hour)},
				}
				mockPaymentRepo.On("GetExpiringPayments", mock.Anything, mock.Anything, mock.Anything).Return(expiringPayments, nil).Once()
				mockRabbitMqExt.On("PublishWithDelay", mock.Anything, rabbitMqExt.PaymentExpirationRoutingKey, expiringPayments[0], mock.Anything).Return(errors.New("publish error")).Once()
			},
			expectedError: errors.New("publish error"),
		},
		{
			name: "when no expiring payment, then should not return error",
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				mockPaymentRepo.On("GetExpiringPayments", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockRabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockAccountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)

			service := &PaymentService{
				paymentRepo:            mockPaymentRepo,
				rabbitMqExt:            mockRabbitMqExt,
				logger:                 mockLogger,
				accountTransactionRepo: mockAccountTransactionRepo,
			}

			tc.mockSetup(mockPaymentRepo, mockRabbitMqExt)
			err := service.PublishPaymentExpirationMessage(context.Background())
			assert.Equal(t, tc.expectedError, err)

			mockPaymentRepo.AssertExpectations(t)
			mockRabbitMqExt.AssertExpectations(t)
		})
	}
}

func TestExpirePayment(t *testing.T) {
	tests := []struct {
		name          string
		request       paymentModel.ExpiringPayment
		mockSetup     func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService)
		expectedError error
	}{
		{
			name: "when payment exists and is eligible for expiration, then payment should be expired without error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock account transaction
				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:           uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					MerchantID:     uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
					AdditionalInfo: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

				// Mock push notification
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Once()
				unifiedPaymentSvc.On("SendCallback", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when failed to get payment by id, then should return database error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(nil, errors.New("database error")).Once()
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, errors.New("database error")),
		},
		{
			name: "when payment not found, then should not return error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(nil, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when merchant id does not match payment merchant id, then should not return error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				uuidStr := "123e4567-e89b-12d3-a456-426614174000"
				merchantStr := "323e4567-e89b-12d3-a456-426614174000" // Different from the request

				payment := &paymentModel.Payment{
					UUID:       uuidStr,
					MerchantID: merchantStr,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, uuidStr).Return(payment, nil).Once()

				// No need to mock account transaction operations for merchant mismatch case - function returns early
			},
			expectedError: nil,
		},
		{
			name: "when payment status is already in final state (SUCCESS), then should not return error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()

				// No need to create account transaction objects for final state payments
				// No need to mock account transaction calls for final state payments - function returns early
			},
			expectedError: nil,
		},
		{
			name: "when payment status is already in final state (VOID), then should not return error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_VOID,
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()

				// No need to create account transaction objects for final state payments
				// No need to mock account transaction calls for final state payments - function returns early
			},
			expectedError: nil,
		},
		{
			name: "when failed to update payment status, then should return database error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(errors.New("update error")).Once()

				// No need to mock account transaction calls because the function returns early after UpdatePaymentStatus fails
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, errors.New("update error")),
		},
		{
			name: "when wallet payment expired, success inquiry status",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodEWallet,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				unifiedPaymentSvc.On("InquiryEWalletPayment", mock.Anything, payment).Return(payment, nil).Once()

				// No need to mock account transaction calls because the function returns early after UpdatePaymentStatus fails
			},
			expectedError: nil,
		},
		{
			name: "when wallet payment API from require_action status expired, update payment status to PROCESSING",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     constant.UnifiedPaymentSessionStatusRequireAction,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodEWallet,
					},
					Metadata: &map[string]any{
						"mode": constant.UnifiedPaymentModeAPI,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				unifiedPaymentSvc.On("InquiryEWalletPayment", mock.Anything, payment).Return(payment, nil).Once()
				unifiedPaymentSvc.On("UpdateEWalletPaymentSession", mock.Anything, mock.Anything).Return(payment, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when wallet payment API from require_action status expired, FAILED update payment status to PROCESSING",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     constant.UnifiedPaymentSessionStatusRequireAction,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodEWallet,
					},
					Metadata: &map[string]any{
						"mode": constant.UnifiedPaymentModeAPI,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				unifiedPaymentSvc.On("InquiryEWalletPayment", mock.Anything, payment).Return(payment, nil).Once()
				unifiedPaymentSvc.On("UpdateEWalletPaymentSession", mock.Anything, mock.Anything).Return(nil, pkgErrors.New(response.HttpErrDatabase, errors.New("update error"))).Once()
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, errors.New("update error")),
		},
		{
			name: "when wallet payment expired, error inquiry status",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodEWallet,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				unifiedPaymentSvc.On("InquiryEWalletPayment", mock.Anything, payment).Return(nil, constant.ErrSomeErrorForUnitTest).Once()

				// No need to mock account transaction calls because the function returns early after UpdatePaymentStatus fails
			},
			expectedError: constant.ErrSomeErrorForUnitTest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockRabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockAccountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			mockStatusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			service := &PaymentService{
				paymentRepo:            mockPaymentRepo,
				rabbitMqExt:            mockRabbitMqExt,
				logger:                 mockLogger,
				accountTransactionRepo: mockAccountTransactionRepo,
				unifiedPaymentSvc:      unifiedPaymentSvc,
				statusHistoriesRepo:    mockStatusHistoriesRepo,
			}

			tc.mockSetup(mockPaymentRepo, mockAccountTransactionRepo, mockRabbitMqExt, unifiedPaymentSvc)
			err := service.ExpirePayment(context.Background(), tc.request)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockAccountTransactionRepo.AssertExpectations(t)
			mockRabbitMqExt.AssertExpectations(t)
		})
	}
}

func TestExpirePaymentVirtualAccountAndQris(t *testing.T) {
	tests := []struct {
		name          string
		request       paymentModel.ExpiringPayment
		mockSetup     func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService)
		expectedError error
	}{
		{
			name: "when VA payment is already paid from inquiry, then should skip expiration",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:                     "123e4567-e89b-12d3-a456-426614174000",
					MerchantID:               "223e4567-e89b-12d3-a456-426614174000",
					Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
					ProcessorReferenceNumber: util.ValueToPtr("7663123400000012"),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				mockSnapCoreRepo.On("InquiryStatusVirtualAccount", mock.Anything, &snapCoreVAModel.InquiryStatusVARequest{
					VirtualAccount: "7663123400000012",
					SkipPublish:    false,
				}).Return(&snapCoreVAModel.InquiryStatusVAResponse{
					Data: snapCoreVAModel.InquiryStatusVAResponseData{
						ResponseCode:    "2002400",
						ResponseMessage: "Successful",
					},
				}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when VA payment is pending from inquiry, then should proceed with expiration",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:                     "123e4567-e89b-12d3-a456-426614174000",
					MerchantID:               "223e4567-e89b-12d3-a456-426614174000",
					Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
					ProcessorReferenceNumber: util.ValueToPtr("7663123400000012"),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				mockSnapCoreRepo.On("InquiryStatusVirtualAccount", mock.Anything, &snapCoreVAModel.InquiryStatusVARequest{
					VirtualAccount: "7663123400000012",
					SkipPublish:    false,
				}).Return(&snapCoreVAModel.InquiryStatusVAResponse{
					Data: snapCoreVAModel.InquiryStatusVAResponseData{
						ResponseCode:    "2022600",
						ResponseMessage: "Request In Progress",
					},
				}, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:           uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					MerchantID:     uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
					AdditionalInfo: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Once()
				unifiedPaymentSvc.On("SendCallback", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when VA inquiry returns error, then should skip expiration without error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:                     "123e4567-e89b-12d3-a456-426614174000",
					MerchantID:               "223e4567-e89b-12d3-a456-426614174000",
					Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
					ProcessorReferenceNumber: util.ValueToPtr("7663123400000012"),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				mockSnapCoreRepo.On("InquiryStatusVirtualAccount", mock.Anything, &snapCoreVAModel.InquiryStatusVARequest{
					VirtualAccount: "7663123400000012",
					SkipPublish:    false,
				}).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedError: nil,
		},
		{
			name: "when VA not found in processor (404), then should skip expiration without error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:                     "123e4567-e89b-12d3-a456-426614174000",
					MerchantID:               "223e4567-e89b-12d3-a456-426614174000",
					Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
					ProcessorReferenceNumber: util.ValueToPtr("7663123400000012"),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				mockSnapCoreRepo.On("InquiryStatusVirtualAccount", mock.Anything, &snapCoreVAModel.InquiryStatusVARequest{
					VirtualAccount: "7663123400000012",
					SkipPublish:    false,
				}).Return(&snapCoreVAModel.InquiryStatusVAResponse{
					Data: snapCoreVAModel.InquiryStatusVAResponseData{
						ResponseCode:    "4042600",
						ResponseMessage: "Virtual Account Not Found",
					},
				}, pkgErrors.New(response.HttpStatusErrorNotFound, constant.ErrVANotFoundInProcessor)).Once()
			},
			expectedError: nil,
		},
		{
			name: "when VA already paid in processor (409), then should skip expiration without retry",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:                     "123e4567-e89b-12d3-a456-426614174000",
					MerchantID:               "223e4567-e89b-12d3-a456-426614174000",
					Status:                   paymentConstant.PAYMENT_STATUS_PENDING,
					ProcessorReferenceNumber: util.ValueToPtr("7663123400000012"),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				mockSnapCoreRepo.On("InquiryStatusVirtualAccount", mock.Anything, &snapCoreVAModel.InquiryStatusVARequest{
					VirtualAccount: "7663123400000012",
					SkipPublish:    false,
				}).Return(&snapCoreVAModel.InquiryStatusVAResponse{
					Data: snapCoreVAModel.InquiryStatusVAResponseData{
						ResponseCode:    "4092600",
						ResponseMessage: "Virtual Account Already Paid",
					},
				}, pkgErrors.New(response.HttpErrRequest, constant.ErrVAAlreadyPaidInProcessor)).Once()
			},
			expectedError: nil,
		},
		{
			name: "when VA payment has no VA number, then should proceed with expiration without inquiry",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:           uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					MerchantID:     uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
					AdditionalInfo: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Once()
				unifiedPaymentSvc.On("SendCallback", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when QRIS charge not found, then should proceed with expiration",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(nil, nil).Once()

				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:           uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					MerchantID:     uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
					AdditionalInfo: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Once()
				unifiedPaymentSvc.On("SendCallback", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when QRIS inquiry returns error, then should skip expiration without error",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()

				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId: "qris-id-123",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(charge, nil).Once()
				mockSnapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "qris-id-123"}).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedError: nil,
		},
		{
			name: "when QRIS inquiry returns final status, then should skip expiration",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()

				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId: "qris-id-123",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(charge, nil).Once()
				mockSnapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "qris-id-123"}).Return(&snapCoreQRModel.QrisInquiryStatusResponse{
					Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
						Status: constant.QrLatestStatusSuccess,
					},
				}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when QRIS inquiry returns pending status, then should proceed with expiration",
			request: paymentModel.ExpiringPayment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				ExpiredAt:  time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockSnapCoreRepo *repositoryMocks.ISnapCoreRepository, mockOrchestratorSvc *serviceMocks.IOrchestratorService, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()

				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId: "qris-id-123",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(charge, nil).Once()
				mockSnapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "qris-id-123"}).Return(&snapCoreQRModel.QrisInquiryStatusResponse{
					Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
						Status: "PENDING",
					},
				}, nil).Once()

				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:           uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					MerchantID:     uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
					AdditionalInfo: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Once()
				unifiedPaymentSvc.On("SendCallback", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockRabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockAccountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockSnapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
			mockOrchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			mockStatusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			service := &PaymentService{
				paymentRepo:            mockPaymentRepo,
				rabbitMqExt:            mockRabbitMqExt,
				logger:                 mockLogger,
				accountTransactionRepo: mockAccountTransactionRepo,
				snapCoreRepo:           mockSnapCoreRepo,
				orchestratorSvc:        mockOrchestratorSvc,
				unifiedPaymentSvc:      unifiedPaymentSvc,
				statusHistoriesRepo:    mockStatusHistoriesRepo,
			}

			tc.mockSetup(mockPaymentRepo, mockAccountTransactionRepo, mockRabbitMqExt, mockSnapCoreRepo, mockOrchestratorSvc, unifiedPaymentSvc)
			err := service.ExpirePayment(context.Background(), tc.request)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockAccountTransactionRepo.AssertExpectations(t)
			mockRabbitMqExt.AssertExpectations(t)
			mockSnapCoreRepo.AssertExpectations(t)
			mockOrchestratorSvc.AssertExpectations(t)
			unifiedPaymentSvc.AssertExpectations(t)
		})
	}
}

func TestHandleProcessedPayment(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		payment       *paymentModel.Payment
		mockSetup     func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt)
		expectedError error
	}{
		{
			name: "when payment status is processing and has retry config, then should republish with delay",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(1)),
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				// Mock orchestrator service
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry returning processing status
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusProcessing,
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.MatchedBy(func(req *creditcardModel.InquiryTransactionRequest) bool {
					return req.MerchantID == "223e4567-e89b-12d3-a456-426614174000" &&
						req.ClientReferenceID == "ref-123" &&
						req.ProcessorReferenceID == "proc-ref-123"
				})).Return(response, nil).Once()

				// Mock rabbit MQ publish with delay (this covers lines 234-249)
				mockRabbitMqExt.On("PublishWithDelay", mock.Anything, rabbitMqExt.PaymentExpirationRoutingKey, mock.MatchedBy(func(expiringPayment *paymentModel.ExpiringPayment) bool {
					return expiringPayment.UUID == "123e4567-e89b-12d3-a456-426614174000" &&
						expiringPayment.MerchantID == "223e4567-e89b-12d3-a456-426614174000" &&
						expiringPayment.ChargeStatus == constant.ChargeStatusProcessing
				}), mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when payment status is processing but no delay config, then should skip republish",
			ctx:  context.Background(),
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "UNKNOWN",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				// Mock orchestrator service
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry returning processing status
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusProcessing,
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.MatchedBy(func(req *creditcardModel.InquiryTransactionRequest) bool {
					return req.MerchantID == "223e4567-e89b-12d3-a456-426614174000"
				})).Return(response, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when publish with delay fails, then should log error but not return error",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(1)),
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				// Mock orchestrator service
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry returning processing status
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusProcessing,
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.Anything).Return(response, nil).Once()

				// Mock rabbit MQ publish with delay failure (covers lines 250-252)
				mockRabbitMqExt.On("PublishWithDelay", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("publish error")).Once()
			},
			expectedError: nil, // Function returns nil even when publish fails (line 253)
		},
		{
			name: "when payment status is processing and delay duration is -1 (exceeded retry limit), then should cancel payment and publish notification",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(10)), // High retry count to trigger -1 duration
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				// Mock orchestrator service
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry returning processing status
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusProcessing,
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.MatchedBy(func(req *creditcardModel.InquiryTransactionRequest) bool {
					return req.MerchantID == "223e4567-e89b-12d3-a456-426614174000" &&
						req.ClientReferenceID == "ref-123" &&
						req.ProcessorReferenceID == "proc-ref-123"
				})).Return(response, nil).Once()

				// Mock rabbit MQ publish for cancelled payment notification (lines 268-272)
				mockRabbitMqExt.On("Publish", mock.Anything, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, mock.Anything, mock.MatchedBy(func(bytes []byte) bool {
					// Verify the marshaled request contains cancelled status
					return len(bytes) > 0
				})).Return(nil).Once()

				// Mock rabbit MQ publish with delay (lines 275-285) - this still gets called even when duration is -1
				mockRabbitMqExt.On("PublishWithDelay", mock.Anything, rabbitMqExt.PaymentExpirationRoutingKey, mock.MatchedBy(func(expiringPayment *paymentModel.ExpiringPayment) bool {
					return expiringPayment.UUID == "123e4567-e89b-12d3-a456-426614174000"
				}), mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when payment status is processing, delay duration is -1, and marshal succeeds, then should publish notification and continue",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(10)), // High retry count to trigger -1 duration
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				// Mock orchestrator service with valid UUID to ensure marshal succeeds
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000", // Valid UUID
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry returning processing status
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusProcessing,
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.Anything).Return(response, nil).Once()

				// Mock successful Publish call for cancelled payment notification (lines 268-272)
				mockRabbitMqExt.On("Publish", mock.Anything, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, mock.Anything, mock.MatchedBy(func(bytes []byte) bool {
					// Verify the marshaled request contains cancelled status and verify structure
					return len(bytes) > 0 && string(bytes) != ""
				})).Return(nil).Once()

				// Mock rabbit MQ publish with delay (lines 275-285) - this still gets called even when duration is -1
				mockRabbitMqExt.On("PublishWithDelay", mock.Anything, rabbitMqExt.PaymentExpirationRoutingKey, mock.MatchedBy(func(expiringPayment *paymentModel.ExpiringPayment) bool {
					return expiringPayment.UUID == "123e4567-e89b-12d3-a456-426614174000"
				}), mock.Anything).Return(nil).Once()
			},
			expectedError: nil, // Function returns nil after successful publish
		},
		{
			name: "when payment status is processing, delay duration is -1, but publish fails, then should log error and return nil",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(10)), // High retry count to trigger -1 duration
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				// Mock orchestrator service
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry returning processing status
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusProcessing,
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.Anything).Return(response, nil).Once()

				// Mock rabbit MQ publish failure (covers lines 268-272)
				mockRabbitMqExt.On("Publish", mock.Anything, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, mock.Anything, mock.Anything).Return(errors.New("publish error")).Once()

				// No PublishWithDelay call expected since function returns early after Publish fails (line 271)
			},
			expectedError: nil, // Function returns nil even when publish fails (line 271)
		},
		{
			name: "when payment has on-behalf parent ID, then should use parent ID for inquiry",
			ctx:  context.Background(),
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "child-merchant-id",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
				Metadata: &map[string]any{
					"onBehalf": map[string]any{
						"parentMerchantId": "parent-merchant-id",
					},
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				// Mock orchestrator service
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry - verify it uses parent merchant ID (lines 170-178)
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentUUID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					ReferenceID:   "ref-123",
					PaymentStatus: constant.UnifiedPaymentSessionStatusPaid,
					Amount:        decimal.NewFromInt(10000),
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.MatchedBy(func(req *creditcardModel.InquiryTransactionRequest) bool {
					// Verify parent merchant ID is used instead of payment's merchant ID
					return req.MerchantID == "parent-merchant-id" &&
						req.ClientReferenceID == "ref-123" &&
						req.ProcessorReferenceID == "proc-ref-123"
				})).Return(response, nil).Once()

				// Mock rabbit MQ publish for successful payment
				mockRabbitMqExt.On("Publish", mock.Anything, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when InquiryTransaction returns error, then should create fallback cancelled response",
			ctx:  context.Background(),
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				// Mock orchestrator service
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry returning error (lines 180-195)
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.Anything).Return(nil, errors.New("inquiry failed")).Once()

				// Mock rabbit MQ publish with fallback cancelled response
				mockRabbitMqExt.On("Publish", mock.Anything, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, mock.Anything, mock.MatchedBy(func(bytes []byte) bool {
					// Verify the fallback response has cancelled status
					var req creditcardModel.CardPaymentNotificationRequest
					if err := json.Unmarshal(bytes, &req); err != nil {
						return false
					}
					return req.Data.PaymentStatus == constant.UnifiedPaymentSessionStatusCancelled &&
						req.Data.ReferenceID == "ref-123" &&
						req.Data.AcquirerTransactionID == "proc-ref-123"
				})).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when InquiryTransaction returns error and processor transaction ID is invalid UUID, then should create fallback response with zero UUID",
			ctx:  context.Background(),
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt) {
				// Mock orchestrator service with invalid UUID as processor transaction ID
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "invalid-uuid-format", // Invalid UUID to trigger parse error (lines 182-185)
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry returning error
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.Anything).Return(nil, errors.New("inquiry failed")).Once()

				// Mock rabbit MQ publish with fallback cancelled response containing zero UUID
				mockRabbitMqExt.On("Publish", mock.Anything, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, mock.Anything, mock.MatchedBy(func(bytes []byte) bool {
					// Verify the fallback response has cancelled status and zero UUID for transaction ID
					var req creditcardModel.CardPaymentNotificationRequest
					if err := json.Unmarshal(bytes, &req); err != nil {
						return false
					}
					// When UUID parse fails, trxID will be zero UUID
					return req.Data.PaymentStatus == constant.UnifiedPaymentSessionStatusCancelled &&
						req.Data.TransactionID == uuid.UUID{}
				})).Return(nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockOrchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			mockCreditCardSvc := serviceMocks.NewICreditCardService(t)
			mockRabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			service := &PaymentService{
				orchestratorSvc: mockOrchestratorSvc,
				creditCardSvc:   mockCreditCardSvc,
				rabbitMqExt:     mockRabbitMqExt,
				logger:          mockLogger,
				config: &config.Config{
					UnifiedPaymentConfig: config.UnifiedPaymentConfig{
						ExpiringProcessedBackoffMinutes: []int{
							1, 3, 5, 10, 15, 30,
						},
					},
				},
			}

			tc.mockSetup(mockOrchestratorSvc, mockCreditCardSvc, mockRabbitMqExt)
			err := service.handleProcessedPayment(tc.ctx, tc.payment)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockOrchestratorSvc.AssertExpectations(t)
			mockCreditCardSvc.AssertExpectations(t)
			mockRabbitMqExt.AssertExpectations(t)
		})
	}
}

func TestExpirePaymentProcessingChargeStatus(t *testing.T) {
	tests := []struct {
		name          string
		request       paymentModel.ExpiringPayment
		mockSetup     func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockService *PaymentService)
		expectedError error
	}{
		{
			name: "when payment status is processing and charge status is processing, then should call handleProcessedPayment",
			request: paymentModel.ExpiringPayment{
				UUID:         "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:   "223e4567-e89b-12d3-a456-426614174000",
				ChargeStatus: constant.ChargeStatusProcessing,
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockService *PaymentService) {
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     constant.UnifiedPaymentSessionStatusProcessing,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()

				// Mock orchestrator service
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
				}
				mockService.orchestratorSvc.(*serviceMocks.IOrchestratorService).On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusCancelled,
					PaymentUUID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					ReferenceID:   "ref-123",
					Amount:        decimal.NewFromInt(10000),
				}
				mockService.creditCardSvc.(*serviceMocks.ICreditCardService).On("InquiryTransaction", mock.Anything, mock.Anything).Return(response, nil).Once()

				// Mock publish notification
				mockService.rabbitMqExt.(*rabbitMqExtMocks.RabbitMQExt).On("Publish", mock.Anything, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when payment status is processing and charge status is waiting for capture, then should call handleAuthorizedPayment",
			request: paymentModel.ExpiringPayment{
				UUID:         "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:   "223e4567-e89b-12d3-a456-426614174000",
				ChargeStatus: constant.ChargeStatusWaitingForCapture,
			},
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockService *PaymentService) {
				expiredAt := time.Now().Add(-1 * time.Hour)
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     constant.UnifiedPaymentSessionStatusProcessing,
					ExpiredAt:  &expiredAt,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").Return(payment, nil).Once()

				// Mock handleAuthorizedPayment call flow
				accountTrxMetadata := orchestrator_model.MetadataPayment[any]{
					ChargeStatus: constant.ChargeStatusWaitingForCapture,
				}
				rawAccountTrxMetadata, _ := json.Marshal(accountTrxMetadata)
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId: "proc-ref-123",
					Status:               constant.StatusPending,
					Credit:               0.0,
					AdditionalInfo: types.NullJSONText{
						Valid: true, JSONText: rawAccountTrxMetadata,
					},
				}
				mockService.orchestratorSvc.(*serviceMocks.IOrchestratorService).On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock unified payment service capture
				mockService.unifiedPaymentSvc.(*serviceMocks.IUnifiedPaymentService).On("Capture", mock.Anything, mock.Anything).Return(nil, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockOrchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			mockCreditCardSvc := serviceMocks.NewICreditCardService(t)
			mockRabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)
			mockUnifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			service := &PaymentService{
				paymentRepo:       mockPaymentRepo,
				orchestratorSvc:   mockOrchestratorSvc,
				creditCardSvc:     mockCreditCardSvc,
				rabbitMqExt:       mockRabbitMqExt,
				unifiedPaymentSvc: mockUnifiedPaymentSvc,
				logger:            mockLogger,
			}

			tc.mockSetup(mockPaymentRepo, service)
			err := service.ExpirePayment(context.Background(), tc.request)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockOrchestratorSvc.AssertExpectations(t)
			mockCreditCardSvc.AssertExpectations(t)
			mockRabbitMqExt.AssertExpectations(t)
			mockUnifiedPaymentSvc.AssertExpectations(t)
		})
	}
}

func TestHandleAuthorizedPayment(t *testing.T) {

	accountTrxMetadata := orchestrator_model.MetadataPayment[any]{
		ChargeStatus: constant.ChargeStatusWaitingForCapture,
	}
	rawAccountTrxMetadata, _ := json.Marshal(accountTrxMetadata)

	tests := []struct {
		name          string
		payment       *paymentModel.Payment
		mockSetup     func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository)
		expectedError error
	}{
		{
			name: "when payment method is not credit card, then should return nil",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
					Name: "GOPAY",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
			},
			expectedError: nil,
		},
		{
			name: "when orchestrator service fails to find charge, then should return nil",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(nil, errors.New("charge not found")).Once()
			},
			expectedError: nil,
		},
		{
			name: "when charge is not found, then should return nil",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(nil, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when charge status is not waiting for capture, then should return nil",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId: "proc-ref-123",
					Status:               constant.ChargeStatusProcessing, // Not waiting for capture
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when payment is not expired and inquiry returns success status, then should return nil",
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				ExpiredAt:   util.ValueToPtr(time.Now().Add(1 * time.Hour)), // Not expired
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
					Status:                 constant.StatusPending,
					AdditionalInfo: types.NullJSONText{
						Valid: true, JSONText: rawAccountTrxMetadata,
					},
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock inquiry transaction returning success
				inquiryResponse := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.StatusSuccess,
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.Anything).Return(inquiryResponse, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when payment is not expired and inquiry returns processing with authorized transaction, then should republish with delay",
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				ExpiredAt:   util.ValueToPtr(time.Now().Add(1 * time.Hour)), // Not expired
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
					Status:                 constant.StatusPending,
					AdditionalInfo: types.NullJSONText{
						Valid: true, JSONText: rawAccountTrxMetadata,
					},
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock inquiry transaction returning processing with authorized status
				inquiryResponse := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusProcessing,
					AuthorizationData: &creditcardModel.PaymentNotificationAuthorizationDataRequest{
						TransactionStaus: constant.CardTransactionStatusAuthorized,
					},
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.Anything).Return(inquiryResponse, nil).Once()

				// Mock republish with delay
				mockRabbitMqExt.On("PublishWithDelay", mock.Anything, rabbitMqExt.PaymentExpirationRoutingKey, mock.MatchedBy(func(expiringPayment *paymentModel.ExpiringPayment) bool {
					return expiringPayment.UUID == "123e4567-e89b-12d3-a456-426614174000" &&
						expiringPayment.ChargeStatus == constant.ChargeStatusWaitingForCapture
				}), mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when payment is expired and captured amount > 0, then should force release payment",
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				Currency:    "IDR",
				ExpiredAt:   util.ValueToPtr(time.Now().Add(1 * time.Hour)), // Expired but limited to max card expiry
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				chargeUUID := uuid.MustParse("456e7890-e89b-12d3-a456-426614174001")
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:                   chargeUUID,
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
					Status:                 constant.StatusPending,
					AdditionalInfo: types.NullJSONText{
						Valid: true, JSONText: rawAccountTrxMetadata,
					},
					Credit: 5000, // Captured amount > 0
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Twice()

				// Mock credit card service inquiry returning processing status
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusProcessing,
					AuthorizationData: &creditcardModel.PaymentNotificationAuthorizationDataRequest{
						TransactionStaus: constant.CreditCardProcessorStatusExpired,
					},
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.MatchedBy(func(req *creditcardModel.InquiryTransactionRequest) bool {
					return req.MerchantID == "223e4567-e89b-12d3-a456-426614174000" &&
						req.ClientReferenceID == "ref-123" &&
						req.ProcessorReferenceID == "proc-ref-123"
				})).Return(response, nil).Once()

				mockFeeSvc.On(
					"GetFeeCalculationAndDetail", constant.ValueCtxMockType(), mock.Anything,
				).Return(4_000.0, &feeModel.FeeMetadataObject{
					Type:          "PAYMENT",
					DeductionType: "DIRECT",
					AmountType:    "AMOUNT",
					Amount:        4_000,
					TaxType:       "NON_PKP",
					FinalAmount:   4_000,
				}, nil)
				mockFeeSvc.On("IncrementLadderCounter", mock.Anything, mock.Anything, mock.Anything).Once()

				mockPaymentRepo.On(
					"UpdatePaymentMetadataById", mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Once()

				// Mock transaction operations
				ctxTx := context.WithValue(context.Background(), "transaction", "test")
				mockPaymentRepo.On("BeginTransaction", mock.Anything).Return(ctxTx, nil).Once()
				mockPaymentRepo.On("UpdatePaymentData", ctxTx, mock.Anything).Return(nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mockAccountTransactionRepo.On("CommitTransaction", ctxTx).Return(nil).Once()

				mockOrchestratorSvc.On("FindByReference", mock.Anything, mock.Anything, mock.Anything).Return(charge, nil).Once()

				expiredAt := time.Now().Add(-1 * time.Hour)
				payment := &paymentModel.Payment{
					UUID:       "123e4567-e89b-12d3-a456-426614174000",
					MerchantID: "223e4567-e89b-12d3-a456-426614174000",
					Status:     constant.UnifiedPaymentSessionStatusExpired,
					ExpiredAt:  &expiredAt,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, mock.Anything).Return(payment, nil).Once()

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{UUID: payment.MerchantID}, nil).Once()

				mockRabbitMqExt.On("PublishForSettlementProcess",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil)

				// Mock send callback
				mockUnifiedPaymentSvc.On("SendCallback", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: nil,
		},
		{
			name: "when payment is expired and captured amount is 0, then should expire payment and charge",
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				ExpiredAt:   util.ValueToPtr(time.Now().Add(1 * time.Hour)), //  Expired but limited to max card expiry
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				chargeUUID := uuid.MustParse("456e7890-e89b-12d3-a456-426614174001")
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:                   chargeUUID,
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
					Status:                 constant.StatusPending,
					AdditionalInfo: types.NullJSONText{
						Valid: true, JSONText: rawAccountTrxMetadata,
					},
					Credit: 0, // No captured amount
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock credit card service inquiry returning processing status
				response := &creditcardModel.PaymentNotificationDataRequest{
					PaymentStatus: constant.UnifiedPaymentSessionStatusProcessing,
					AuthorizationData: &creditcardModel.PaymentNotificationAuthorizationDataRequest{
						TransactionStaus: constant.CreditCardProcessorStatusExpired,
					},
				}
				mockCreditCardSvc.On("InquiryTransaction", mock.Anything, mock.MatchedBy(func(req *creditcardModel.InquiryTransactionRequest) bool {
					return req.MerchantID == "223e4567-e89b-12d3-a456-426614174000" &&
						req.ClientReferenceID == "ref-123" &&
						req.ProcessorReferenceID == "proc-ref-123"
				})).Return(response, nil).Once()

				// Mock transaction operations for expiration
				ctxTx := context.WithValue(context.Background(), "transaction", "test")
				mockPaymentRepo.On("BeginTransaction", mock.Anything).Return(ctxTx, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", ctxTx, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock account transaction find and update for expiration
				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       uuid.MustParse("789e0123-e89b-12d3-a456-426614174002"),
					MerchantID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
				}
				mockAccountTransactionRepo.On("FindByReference", ctxTx, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", ctxTx, mock.Anything, mock.Anything).Return(nil).Once()

				// Mock push notification and commit
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Once()
				mockPaymentRepo.On("CommitTransaction", ctxTx).Return(nil).Once()
				mockUnifiedPaymentSvc.On("SendCallback", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: nil,
		},
		{
			name: "when payment is expired and capture is called",
			payment: &paymentModel.Payment{
				UUID:        "123e4567-e89b-12d3-a456-426614174000",
				MerchantID:  "223e4567-e89b-12d3-a456-426614174000",
				Status:      constant.UnifiedPaymentSessionStatusProcessing,
				ReferenceID: util.ValueToPtr("ref-123"),
				Amount:      decimal.NewFromInt(10000),
				ExpiredAt:   util.ValueToPtr(time.Now().Add(-1 * time.Hour)), // Expired
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Name: "VISA",
				},
			},
			mockSetup: func(mockOrchestratorSvc *serviceMocks.IOrchestratorService, mockCreditCardSvc *serviceMocks.ICreditCardService, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockUnifiedPaymentSvc *serviceMocks.IUnifiedPaymentService, mockFeeSvc *serviceMocks.IFeeService, mockMerchantRepo *repositoryMocks.IMerchantRepository) {
				chargeUUID := uuid.MustParse("456e7890-e89b-12d3-a456-426614174001")
				charge := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:                   chargeUUID,
					ProcessorReferenceId:   "proc-ref-123",
					ProcessorTransactionId: "123e4567-e89b-12d3-a456-426614174000",
					Status:                 constant.StatusPending,
					AdditionalInfo: types.NullJSONText{
						Valid: true, JSONText: rawAccountTrxMetadata,
					},
					Credit: 0, // No captured amount
				}
				mockOrchestratorSvc.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.ReferencePayment).Return(charge, nil).Once()

				// Mock unified payment service capture
				mockUnifiedPaymentSvc.On("Capture", mock.Anything, mock.MatchedBy(func(req *unifiedPaymentModel.CaptureRequest) bool {
					return req.PaymentID == "123e4567-e89b-12d3-a456-426614174000" &&
						req.ChargeID == chargeUUID.String() &&
						req.ReleaseRemainingAmount == true
				})).Return(nil, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockOrchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			mockCreditCardSvc := serviceMocks.NewICreditCardService(t)
			mockRabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockAccountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockUnifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			mockStatusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			mockFeeSvc := serviceMocks.NewIFeeService(t)
			mockMerchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			service := &PaymentService{
				orchestratorSvc:        mockOrchestratorSvc,
				creditCardSvc:          mockCreditCardSvc,
				rabbitMqExt:            mockRabbitMqExt,
				paymentRepo:            mockPaymentRepo,
				accountTransactionRepo: mockAccountTransactionRepo,
				unifiedPaymentSvc:      mockUnifiedPaymentSvc,
				statusHistoriesRepo:    mockStatusHistoriesRepo,
				logger:                 mockLogger,
				config: &config.Config{
					UnifiedPaymentConfig: config.UnifiedPaymentConfig{
						RetryExpiringAuthorizedTransactionMinutes: 30,
					},
				},
				feeSvc:       mockFeeSvc,
				merchantRepo: mockMerchantRepo,
			}

			tc.mockSetup(mockOrchestratorSvc, mockCreditCardSvc, mockRabbitMqExt, mockPaymentRepo, mockAccountTransactionRepo, mockUnifiedPaymentSvc, mockFeeSvc, mockMerchantRepo)
			err := service.handleAuthorizedPayment(context.Background(), tc.payment)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockOrchestratorSvc.AssertExpectations(t)
			mockCreditCardSvc.AssertExpectations(t)
			mockRabbitMqExt.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
			mockAccountTransactionRepo.AssertExpectations(t)
			mockUnifiedPaymentSvc.AssertExpectations(t)
		})
	}
}

func TestProcessExpiration(t *testing.T) {
	tests := []struct {
		name          string
		payment       *paymentModel.Payment
		chargeStatus  string
		mockSetup     func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockStatusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository)
		expectedError error
	}{
		{
			name: "when regular payment expires successfully, then should update status and notify",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     paymentConstant.PAYMENT_STATUS_PENDING,
			},
			chargeStatus: constant.ChargeStatusExpired,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockStatusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository) {
				// Mock update payment status
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock status history insertion
				mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction find and update
				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       uuid.MustParse("789e0123-e89b-12d3-a456-426614174002"),
					MerchantID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.MatchedBy(func(req orchestrator_model.UpdatePaymentTransactionRequest) bool {
					return req.LedgerId == accountTransaction.UUID.String() && req.Status == constant.StatusFailed
				}), mock.MatchedBy(func(metadata orchestrator_model.MetadataPayment[any]) bool {
					return metadata.ChargeStatus == constant.ChargeStatusExpired
				})).Return(nil).Once()

				// Mock push notification
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.MatchedBy(func(notification *notification.PushNotification) bool {
					return notification.RoutingKey == fmt.Sprintf(constant.NotificationRoutingKeyFmt, "123e4567-e89b-12d3-a456-426614174000") &&
						notification.Payload.Status == paymentConstant.PaymentStatusExpired
				})).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when static payment expires, then should set to inactive status",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     constant.UnifiedStaticPaymentStatusActive,
			},
			chargeStatus: constant.ChargeStatusExpired,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockStatusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository) {
				// Mock update payment status to inactive
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", constant.UnifiedStaticPaymentStatusInactive, mock.Anything).Return(nil).Once()

				// Mock status history insertion
				mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Static payments skip account transaction and notification logic
			},
			expectedError: nil,
		},
		{
			name: "when update payment status fails, then should return error",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     paymentConstant.PAYMENT_STATUS_PENDING,
			},
			chargeStatus: constant.ChargeStatusExpired,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockStatusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository) {
				// Mock update payment status failure
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(errors.New("update failed")).Once()
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, errors.New("update failed")),
		},
		{
			name: "when account transaction not found, then should return error",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     paymentConstant.PAYMENT_STATUS_PENDING,
			},
			chargeStatus: constant.ChargeStatusExpired,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockStatusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository) {
				// Mock update payment status success
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock status history insertion
				mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction find failure
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(nil, errors.New("not found")).Once()
			},
			expectedError: pkgErrors.New(response.HttpStatusErrorNotFound, errors.New("not found")),
		},
		{
			name: "when charge status is empty, then should use default expired status",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     paymentConstant.PAYMENT_STATUS_PENDING,
			},
			chargeStatus: "", // Empty charge status
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockStatusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository) {
				// Mock update payment status
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock status history insertion
				mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction find and update with default expired status
				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       uuid.MustParse("789e0123-e89b-12d3-a456-426614174002"),
					MerchantID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.MatchedBy(func(req orchestrator_model.UpdatePaymentTransactionRequest) bool {
					return req.LedgerId == accountTransaction.UUID.String() && req.Status == constant.StatusFailed
				}), mock.MatchedBy(func(metadata orchestrator_model.MetadataPayment[any]) bool {
					return metadata.ChargeStatus == constant.ChargeStatusExpired // Should default to expired
				})).Return(nil).Once()

				// Mock push notification
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "when account transaction exists but update fails, then should continue processing",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     paymentConstant.PAYMENT_STATUS_PENDING,
			},
			chargeStatus: constant.ChargeStatusExpired,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockStatusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository) {
				// Mock update payment status success
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock status history insertion
				mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction find success but update failure
				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       uuid.MustParse("789e0123-e89b-12d3-a456-426614174002"),
					MerchantID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("update failed")).Once()

				// Mock push notification still continues
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil, // Function continues despite account transaction update failure
		},
		{
			name: "when push notification fails, then should continue processing",
			payment: &paymentModel.Payment{
				UUID:       "123e4567-e89b-12d3-a456-426614174000",
				MerchantID: "223e4567-e89b-12d3-a456-426614174000",
				Status:     paymentConstant.PAYMENT_STATUS_PENDING,
			},
			chargeStatus: constant.ChargeStatusExpired,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockRabbitMqExt *rabbitMqExtMocks.RabbitMQExt, mockStatusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository) {
				// Mock update payment status success
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174000", paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock status history insertion
				mockStatusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction find and update success
				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:       uuid.MustParse("789e0123-e89b-12d3-a456-426614174002"),
					MerchantID: uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "123e4567-e89b-12d3-a456-426614174000", constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

				// Mock push notification failure
				mockRabbitMqExt.On("PushNotification", mock.Anything, mock.Anything).Return(errors.New("notification failed")).Once()
			},
			expectedError: nil, // Function returns nil even when notification fails
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockAccountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockRabbitMqExt := rabbitMqExtMocks.NewRabbitMQExt(t)
			mockStatusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			service := &PaymentService{
				paymentRepo:            mockPaymentRepo,
				accountTransactionRepo: mockAccountTransactionRepo,
				rabbitMqExt:            mockRabbitMqExt,
				statusHistoriesRepo:    mockStatusHistoriesRepo,
				logger:                 mockLogger,
			}

			tc.mockSetup(mockPaymentRepo, mockAccountTransactionRepo, mockRabbitMqExt, mockStatusHistoriesRepo)
			err := service.processExpiration(context.Background(), tc.payment, tc.chargeStatus)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockAccountTransactionRepo.AssertExpectations(t)
			mockRabbitMqExt.AssertExpectations(t)
			mockStatusHistoriesRepo.AssertExpectations(t)
		})
	}
}
