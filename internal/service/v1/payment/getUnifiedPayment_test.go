package paymentService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestPaymentServiceGetPaymentByReferenceId(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(1 * time.Hour)
	pastTime := now.Add(-1 * time.Hour)
	amount := decimal.NewFromFloat(10000)
	paymentUuid := uuid.NewString()
	referenceId := "REF123"
	merchantID := "merchant-123"
	customerID := "customer-123"

	testCases := []struct {
		name        string
		referenceId string
		merchantID  string
		mockSetup   func(
			paymentRepo *repositoryMocks.IPaymentRepository,
			paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
			customerRepo *repositoryMocks.ICustomerRepository,
			statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
			orchestratorSvc *serviceMocks.IOrchestratorService,
		)
		expectedError string
		wantErr       bool
	}{
		{
			name:        "SUCCESS: Get payment by reference ID",
			referenceId: referenceId,
			merchantID:  merchantID,
			mockSetup: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
				customerRepo *repositoryMocks.ICustomerRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				payment := &paymentModel.Payment{
					UUID:            paymentUuid,
					MerchantID:      merchantID,
					CustomerID:      customerID,
					PaymentMethodID: "pm-123",
					Currency:        "IDR",
					Amount:          amount,
					Status:          paymentConstant.PAYMENT_STATUS_SUCCESS,
					ExpiredAt:       &futureTime,
					CreatedAt:       now,
					UpdatedAt:       now,
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     "QRIS",
						Name:     "QRIS",
						Acquirer: "LinkAja",
					},
				}

				paymentMethod := &paymentModel.PaymentMethod{
					UUID:     "pm-123",
					Type:     "QRIS",
					Name:     "QRIS",
					Acquirer: "LinkAja",
				}

				customer := &customerModel.Customer{
					UUID:        customerID,
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john.doe@example.com",
					PhoneNumber: "+6281234567890",
				}

				paymentItems := []*paymentModel.PaymentItem{
					{
						UUID:        "item-123",
						PaymentID:   paymentUuid,
						Name:        "Test Item",
						Description: "Test Description",
						Qty:         1,
						Currency:    "IDR",
						Amount:      amount,
					},
				}

				accountTransaction := &orchestratorModel.AccountTransactionWithUseCase{
					UUID:           uuid.New(),
					ReferenceID:    paymentUuid,
					Credit:         10000.0,
					Currency:       "IDR",
					Status:         constant.StatusSuccess,
					AdditionalInfo: types.NullJSONText{Valid: true, JSONText: []byte(`{"test": "data"}`)},
				}

				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, referenceId).Return(payment, nil)
				paymentMethodRepo.On("GetPaymentMethodById", mock.Anything, "pm-123").Return(paymentMethod, nil)
				customerRepo.On("FindCustomerById", mock.Anything, customerID).Return(customer, nil)
				paymentRepo.On("GetPaymentItemsByPaymentId", mock.Anything, paymentUuid).Return(paymentItems, nil)
				orchestratorSvc.On("FindByReference", mock.Anything, paymentUuid, constant.TypePayment).Return(accountTransaction, nil)
			},
			wantErr: false,
		},
		{
			name:        "ERROR: Payment not found in repository",
			referenceId: referenceId,
			merchantID:  merchantID,
			mockSetup: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
				customerRepo *repositoryMocks.ICustomerRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, referenceId).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
			wantErr:       true,
		},
		{
			name:        "ERROR: Payment not found (nil result)",
			referenceId: referenceId,
			merchantID:  merchantID,
			mockSetup: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
				customerRepo *repositoryMocks.ICustomerRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, referenceId).Return(nil, nil)
			},
			expectedError: "payment not found",
			wantErr:       true,
		},
		{
			name:        "SUCCESS: Payment expired and status updated",
			referenceId: referenceId,
			merchantID:  merchantID,
			mockSetup: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
				customerRepo *repositoryMocks.ICustomerRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				payment := &paymentModel.Payment{
					UUID:            paymentUuid,
					MerchantID:      merchantID,
					CustomerID:      "",
					PaymentMethodID: "pm-123",
					Currency:        "IDR",
					Amount:          amount,
					Status:          paymentConstant.UnifiedPaymentStatusWaitingForPayment,
					ExpiredAt:       &pastTime,
					CreatedAt:       now,
					UpdatedAt:       now,
					PaymentMethod: paymentModel.PaymentMethod{
						Type:     "QRIS",
						Name:     "QRIS",
						Acquirer: "LinkAja",
					},
				}

				paymentMethod := &paymentModel.PaymentMethod{
					UUID: "pm-123",
					Type: "QRIS",
					Name: "QRIS",
				}

				paymentItems := []*paymentModel.PaymentItem{}

				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, referenceId).Return(payment, nil)
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentUuid, merchantID, paymentConstant.UnifiedPaymentStatusExpired, mock.AnythingOfType("time.Time")).Return(nil)
				statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)
				paymentMethodRepo.On("GetPaymentMethodById", mock.Anything, "pm-123").Return(paymentMethod, nil)
				paymentRepo.On("GetPaymentItemsByPaymentId", mock.Anything, paymentUuid).Return(paymentItems, nil)
				orchestratorSvc.On("FindByReference", mock.Anything, paymentUuid, constant.TypePayment).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name:        "ERROR: Merchant ID mismatch",
			referenceId: referenceId,
			merchantID:  merchantID,
			mockSetup: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
				customerRepo *repositoryMocks.ICustomerRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				payment := &paymentModel.Payment{
					UUID:            paymentUuid,
					MerchantID:      "different-merchant-id",
					CustomerID:      "",
					PaymentMethodID: "pm-123",
					Currency:        "IDR",
					Amount:          amount,
					Status:          paymentConstant.PAYMENT_STATUS_SUCCESS,
					ExpiredAt:       &futureTime,
					CreatedAt:       now,
					UpdatedAt:       now,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: "QRIS",
						Name: "QRIS",
					},
				}

				paymentMethod := &paymentModel.PaymentMethod{
					UUID: "pm-123",
					Type: "QRIS",
					Name: "QRIS",
				}

				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, referenceId).Return(payment, nil)
				paymentMethodRepo.On("GetPaymentMethodById", mock.Anything, "pm-123").Return(paymentMethod, nil)
			},
			expectedError: "payment not found",
			wantErr:       true,
		},
		{
			name:        "ERROR: Payment method not found",
			referenceId: referenceId,
			merchantID:  merchantID,
			mockSetup: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
				customerRepo *repositoryMocks.ICustomerRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				payment := &paymentModel.Payment{
					UUID:            paymentUuid,
					MerchantID:      merchantID,
					CustomerID:      "",
					PaymentMethodID: "pm-123",
					Currency:        "IDR",
					Amount:          amount,
					Status:          paymentConstant.PAYMENT_STATUS_SUCCESS,
					ExpiredAt:       &futureTime,
					CreatedAt:       now,
					UpdatedAt:       now,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: "QRIS",
						Name: "QRIS",
					},
				}

				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, referenceId).Return(payment, nil)
				paymentMethodRepo.On("GetPaymentMethodById", mock.Anything, "pm-123").Return(nil, errors.New("payment method not found"))
			},
			expectedError: "payment method not found",
			wantErr:       true,
		},
		{
			name:        "ERROR: Customer not found",
			referenceId: referenceId,
			merchantID:  merchantID,
			mockSetup: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
				customerRepo *repositoryMocks.ICustomerRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				payment := &paymentModel.Payment{
					UUID:            paymentUuid,
					MerchantID:      merchantID,
					CustomerID:      customerID,
					PaymentMethodID: "pm-123",
					Currency:        "IDR",
					Amount:          amount,
					Status:          paymentConstant.PAYMENT_STATUS_SUCCESS,
					ExpiredAt:       &futureTime,
					CreatedAt:       now,
					UpdatedAt:       now,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: "QRIS",
						Name: "QRIS",
					},
				}

				paymentMethod := &paymentModel.PaymentMethod{
					UUID: "pm-123",
					Type: "QRIS",
					Name: "QRIS",
				}

				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, referenceId).Return(payment, nil)
				paymentMethodRepo.On("GetPaymentMethodById", mock.Anything, "pm-123").Return(paymentMethod, nil)
				customerRepo.On("FindCustomerById", mock.Anything, customerID).Return(nil, errors.New("customer not found"))
			},
			expectedError: "customer not found",
			wantErr:       true,
		},
		{
			name:        "ERROR: Payment items not found",
			referenceId: referenceId,
			merchantID:  merchantID,
			mockSetup: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
				customerRepo *repositoryMocks.ICustomerRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				payment := &paymentModel.Payment{
					UUID:            paymentUuid,
					MerchantID:      merchantID,
					CustomerID:      "",
					PaymentMethodID: "pm-123",
					Currency:        "IDR",
					Amount:          amount,
					Status:          paymentConstant.PAYMENT_STATUS_SUCCESS,
					ExpiredAt:       &futureTime,
					CreatedAt:       now,
					UpdatedAt:       now,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: "QRIS",
						Name: "QRIS",
					},
				}

				paymentMethod := &paymentModel.PaymentMethod{
					UUID: "pm-123",
					Type: "QRIS",
					Name: "QRIS",
				}

				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, referenceId).Return(payment, nil)
				paymentMethodRepo.On("GetPaymentMethodById", mock.Anything, "pm-123").Return(paymentMethod, nil)
				paymentRepo.On("GetPaymentItemsByPaymentId", mock.Anything, paymentUuid).Return(nil, errors.New("payment items not found"))
			},
			expectedError: "payment items not found",
			wantErr:       true,
		},
		{
			name:        "ERROR: Orchestrator service error",
			referenceId: referenceId,
			merchantID:  merchantID,
			mockSetup: func(
				paymentRepo *repositoryMocks.IPaymentRepository,
				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
				customerRepo *repositoryMocks.ICustomerRepository,
				statusHistoriesRepo *repositoryMocks.IStatusHistoriesRepository,
				orchestratorSvc *serviceMocks.IOrchestratorService,
			) {
				payment := &paymentModel.Payment{
					UUID:            paymentUuid,
					MerchantID:      merchantID,
					CustomerID:      "",
					PaymentMethodID: "pm-123",
					Currency:        "IDR",
					Amount:          amount,
					Status:          paymentConstant.PAYMENT_STATUS_SUCCESS,
					ExpiredAt:       &futureTime,
					CreatedAt:       now,
					UpdatedAt:       now,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: "QRIS",
						Name: "QRIS",
					},
				}

				paymentMethod := &paymentModel.PaymentMethod{
					UUID: "pm-123",
					Type: "QRIS",
					Name: "QRIS",
				}

				paymentItems := []*paymentModel.PaymentItem{}

				paymentRepo.On("GetPaymentByMerchantAndReferenceId", mock.Anything, merchantID, referenceId).Return(payment, nil)
				paymentMethodRepo.On("GetPaymentMethodById", mock.Anything, "pm-123").Return(paymentMethod, nil)
				paymentRepo.On("GetPaymentItemsByPaymentId", mock.Anything, paymentUuid).Return(paymentItems, nil)
				orchestratorSvc.On("FindByReference", mock.Anything, paymentUuid, constant.TypePayment).Return(nil, errors.New("orchestrator service error"))
			},
			expectedError: "orchestrator service error",
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)
			customerRepo := repositoryMocks.NewICustomerRepository(t)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(paymentRepo, paymentMethodRepo, customerRepo, statusHistoriesRepo, orchestratorSvc)

			service := &PaymentService{
				paymentRepo:         paymentRepo,
				paymentMethodRepo:   paymentMethodRepo,
				customerRepo:        customerRepo,
				statusHistoriesRepo: statusHistoriesRepo,
				orchestratorSvc:     orchestratorSvc,
				logger:              logger,
			}

			result, err := service.GetPaymentByReferenceId(context.Background(), tc.referenceId, tc.merchantID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				if tc.expectedError != "" {
					assert.Contains(t, err.Error(), tc.expectedError)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, paymentUuid, result.UUID)
				assert.Equal(t, tc.merchantID, result.MerchantID)
				assert.Equal(t, tc.referenceId, result.ReferenceID)
				assert.NotNil(t, result.Amount)
				assert.NotNil(t, result.PaidAmount)
			}

			paymentRepo.AssertExpectations(t)
			paymentMethodRepo.AssertExpectations(t)
			customerRepo.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
		})
	}
}
