package unifiedPaymentService_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/ewallet"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	rabbitMqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInquiryEWalletPayment(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	merchantID := uuid.NewString()
	paymentID := uuid.NewString()
	processorRefID := uuid.NewString()

	metadata := unifiedPaymentModel.MetadataUnifiedPayment{
		IsUnifiedPaymentV2: true,
		PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
			Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
				Channel: c.UnifiedPaymentEWalletDanaAcquirer,
			},
		},
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{},
	}
	metadataBytes, _ := json.Marshal(metadata)
	var metadataInterface map[string]any
	_ = json.Unmarshal(metadataBytes, &metadataInterface)

	shopeeMetadata := unifiedPaymentModel.MetadataUnifiedPayment{
		PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
			Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
				Channel: c.UnifiedPaymentEWalletShopeePayAcquirer,
			},
		},
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{},
	}
	shopeeMetadataBytes, _ := json.Marshal(shopeeMetadata)
	var shopeeMetadataInterface map[string]any
	_ = json.Unmarshal(shopeeMetadataBytes, &shopeeMetadataInterface)

	testCases := []struct {
		name      string
		payment   *paymentModel.Payment
		wantErr   bool
		setupMock func(*repositoryMock.IPaymentRepository, *repositoryMock.IAccountTransactionRepository, *repositoryMock.ISnapCoreRepository, *rabbitMqExtMock.RabbitMQExt)
	}{
		{
			name: "SUCCESS: Skip inquiry for final status",
			payment: &paymentModel.Payment{
				UUID:       paymentID,
				MerchantID: merchantID,
				Status:     c.UnifiedPaymentSessionStatusPaid,
				Metadata:   &metadataInterface,
			},
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				// No mocks needed as it returns early
			},
		},
		{
			name: "ERROR: Payment ledger not found in database",
			payment: &paymentModel.Payment{
				UUID:       paymentID,
				MerchantID: merchantID,
				Status:     c.UnifiedPaymentSessionStatusProcessing,
				Metadata:   &metadataInterface,
			},
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				accountTrxRepo.On("FindByReference",
					mock.Anything,
					paymentID,
					c.TypePayment,
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "ERROR: Payment ledger is nil",
			payment: &paymentModel.Payment{
				UUID:       paymentID,
				MerchantID: merchantID,
				Status:     c.UnifiedPaymentSessionStatusProcessing,
				Metadata:   &metadataInterface,
			},
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				accountTrxRepo.On("FindByReference",
					mock.Anything,
					paymentID,
					c.TypePayment,
				).Return(nil, nil)
			},
		},
		{
			name: "ERROR: SNAP Core inquiry fails",
			payment: &paymentModel.Payment{
				UUID:       paymentID,
				MerchantID: merchantID,
				Status:     c.UnifiedPaymentSessionStatusProcessing,
				Metadata:   &metadataInterface,
			},
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				paymentLedger := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 uuid.New(),
					ProcessorReferenceId: processorRefID,
				}
				accountTrxRepo.On("FindByReference",
					mock.Anything,
					paymentID,
					c.TypePayment,
				).Return(paymentLedger, nil)

				snapCoreRepo.On("InquiryStatusEWalletPayment",
					mock.Anything,
					&ewallet.EWalletInquiryStatusRequest{
						TransactionID: processorRefID,
					},
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "SUCCESS: Payment still processing",
			payment: &paymentModel.Payment{
				UUID:       paymentID,
				MerchantID: merchantID,
				Status:     c.UnifiedPaymentSessionStatusProcessing,
				Metadata:   &metadataInterface,
			},
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, snapCoreRepo *repositoryMock.ISnapCoreRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				paymentLedger := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:                 uuid.New(),
					ProcessorReferenceId: processorRefID,
				}
				accountTrxRepo.On("FindByReference",
					mock.Anything,
					paymentID,
					c.TypePayment,
				).Return(paymentLedger, nil).Once()

				snapCoreRepo.On("InquiryStatusEWalletPayment",
					mock.Anything,
					&ewallet.EWalletInquiryStatusRequest{
						TransactionID: processorRefID,
					},
				).Return(&ewallet.EWalletInquiryStatusResponse{
					LatestTransactionStatus: c.SnapLatestTransactionStatusProcessing[0],
				}, nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			snapCoreRepo := repositoryMock.NewISnapCoreRepository(t)
			rabbitMqMock := rabbitMqExtMock.NewRabbitMQExt(t)

			svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
				WithSnapCoreRepo(snapCoreRepo),
				WithRabbitMQClient(rabbitMqMock),
			)
			tc.setupMock(paymentRepo, accountTrxRepo, snapCoreRepo, rabbitMqMock)

			result, err := svc.InquiryEWalletPayment(context.Background(), tc.payment)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestUpdateEWalletPaymentSession(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	merchantID := uuid.NewString()
	paymentID := uuid.NewString()

	testCases := []struct {
		name      string
		paymentID string
		wantErr   bool
		setupMock func(*repositoryMock.IPaymentRepository, *repositoryMock.IAccountTransactionRepository, *rabbitMqExtMock.RabbitMQExt)
	}{
		{
			name:      "ERROR: Payment not found in database",
			paymentID: paymentID,
			wantErr:   true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById",
					mock.Anything,
					paymentID,
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: Payment is nil",
			paymentID: paymentID,
			wantErr:   true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById",
					mock.Anything,
					paymentID,
				).Return(nil, nil)
			},
		},
		{
			name:      "ERROR: Get payment ledger",
			paymentID: paymentID,
			wantErr:   true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusProcessing,
				}
				paymentRepo.On("GetPaymentById",
					mock.Anything,
					paymentID,
				).Return(payment, nil)

				accountTrxRepo.On("FindByReference",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "SUCCESS: Not Update Payment Status due to Not Require Action",
			paymentID: paymentID,
			wantErr:   false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				metadata := unifiedPaymentModel.MetadataUnifiedPayment{
					IsUnifiedPaymentV2: true,
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: "ewallet",
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
							Channel: c.UnifiedPaymentEWalletDanaAcquirer,
						},
					},
				}
				metadataBytes, _ := json.Marshal(metadata)
				var metadataInterface map[string]any
				_ = json.Unmarshal(metadataBytes, &metadataInterface)

				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusProcessing,
					Metadata:   &metadataInterface,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: "ewallet",
					},
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}
				paymentRepo.On("GetPaymentById",
					mock.Anything,
					paymentID,
				).Return(payment, nil)

				accountTrxRepo.On("FindByReference",
					mock.Anything,
					paymentID,
					c.TypePayment,
				).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID: uuid.New(),
				}, nil)

			},
		},
		{
			name:      "SUCCESS: Update payment status from require action to processing",
			paymentID: paymentID,
			wantErr:   false,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				metadata := unifiedPaymentModel.MetadataUnifiedPayment{
					IsUnifiedPaymentV2: true,
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: "ewallet",
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
							Channel: c.UnifiedPaymentEWalletDanaAcquirer,
						},
					},
				}
				metadataBytes, _ := json.Marshal(metadata)
				var metadataInterface map[string]any
				_ = json.Unmarshal(metadataBytes, &metadataInterface)

				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					Metadata:   &metadataInterface,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: "ewallet",
					},
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}
				paymentRepo.On("GetPaymentById",
					mock.Anything,
					paymentID,
				).Return(payment, nil)

				paymentRepo.On("UpdatePaymentStatus",
					mock.Anything,
					paymentID,
					merchantID,
					c.UnifiedPaymentSessionStatusProcessing,
					mock.Anything,
				).Return(nil)

				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				accountTrxRepo.On("FindByReference",
					mock.Anything,
					paymentID,
					c.TypePayment,
				).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID: uuid.New(),
				}, nil).Maybe()

				rabbitMqMock.On("PublishMerchantCallback", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
				rabbitMqMock.On("PushNotification", mock.Anything, mock.Anything).Return(nil).Maybe()
			},
		},
		{
			name:      "ERROR: Failed to update payment status",
			paymentID: paymentID,
			wantErr:   true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
				}
				paymentRepo.On("GetPaymentById",
					mock.Anything,
					paymentID,
				).Return(payment, nil)

				accountTrxRepo.On("FindByReference",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID: uuid.New(),
				}, nil)

				paymentRepo.On("UpdatePaymentStatus",
					mock.Anything,
					paymentID,
					merchantID,
					c.UnifiedPaymentSessionStatusProcessing,
					mock.Anything,
				).Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: Failed to update payment ledger",
			paymentID: paymentID,
			wantErr:   true,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, rabbitMqMock *rabbitMqExtMock.RabbitMQExt) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
				}
				paymentRepo.On("GetPaymentById",
					mock.Anything,
					paymentID,
				).Return(payment, nil)

				accountTrxRepo.On("FindByReference",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID: uuid.New(),
				}, nil)

				paymentRepo.On("UpdatePaymentStatus",
					mock.Anything,
					paymentID,
					merchantID,
					c.UnifiedPaymentSessionStatusProcessing,
					mock.Anything,
				).Return(nil)

				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(c.ErrSomeErrorForUnitTest)

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			snapCoreRepo := repositoryMock.NewISnapCoreRepository(t)
			rabbitMqMock := rabbitMqExtMock.NewRabbitMQExt(t)

			svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
				WithSnapCoreRepo(snapCoreRepo),
				WithRabbitMQClient(rabbitMqMock),
			)
			tc.setupMock(paymentRepo, accountTrxRepo, rabbitMqMock)

			result, err := svc.UpdateEWalletPaymentSession(context.Background(), tc.paymentID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if result != nil && result.Status == c.UnifiedPaymentSessionStatusProcessing {
					assert.Equal(t, c.UnifiedPaymentSessionStatusProcessing, result.Status)
				}
			}
		})
	}
}
