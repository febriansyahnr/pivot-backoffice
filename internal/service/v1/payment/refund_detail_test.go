package paymentService

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	constant "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestRefundDetailIntegration(t *testing.T) {
	var (
		mockLogger, _              = loggerMocks.NewZapLogger(loggerMocks.Config{})
		mockPaymentRepo            = repositoryMocks.NewIPaymentRepository(t)
		mockAccountTransactionRepo = repositoryMocks.NewIAccountTransactionRepository(t)
		mockCustomerRepo           = repositoryMocks.NewICustomerRepository(t)
		mockStatusHistoriesRepo    = repositoryMocks.NewIStatusHistoriesRepository(t)
		transferSvc                = serviceMocks.NewITransferService(t)
		refundSvc                  = serviceMocks.NewIRefundService(t)
	)

	now := time.Now()

	tests := []struct {
		name         string
		setupMock    func()
		expectRefund bool
		expectError  bool
	}{
		{
			name: "should fetch refund detail when refundId exists in metadata",
			setupMock: func() {
				refundId := "test-refund-id"
				metadataWithRefund := map[string]any{
					"refundId": refundId,
				}

				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "payment-id", "merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "payment-id",
						MerchantID: "merchant-id",
						CustomerID: "customer-id",
						Amount:     decimal.NewFromFloat(100000),
						Currency:   "IDR",
						Status:     constant.StatusSuccess,
						CreatedAt:  now,
						UpdatedAt:  now,
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						},
						Metadata: &metadataWithRefund,
					}, nil).Once()

				mockPaymentRepo.On("GetChargeList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{},
				}, nil).Once()

				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "payment-id", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:                 uuid.New(),
						AccountID:            uuid.New(),
						ReferenceID:          "payment-id",
						TransactionTimestamp: now,
						Credit:               100000,
						Currency:             "IDR",
						Status:               constant.StatusSuccess,
					}, nil).Once()

				// mock status history
				mockStatusHistoriesRepo.On("GetByReference", mock.Anything, constant.TypePayment, "payment-id").
					Return([]*statusHistoryModel.StatusHistory{}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "merchant-id").Return(&customerModel.Customer{
					UUID: "customer-id",
				}, nil).Once()

				// Mock successful refund service call
				refundSvc.On("GetExistingRefundList", mock.Anything, mock.Anything).Return([]refundModel.RefundResponse{
					{
						ID:                refundId,
						ClientReferenceID: "refund-ref-123",
						PaymentSessionID:  "payment-id",
						Status:            constant.StatusSuccess,
						Amount: commonModel.Amount{
							Value:    "50000.00",
							Currency: "IDR",
						},
						CreatedAt: now,
						UpdatedAt: now,
					},
				}, nil).Once()
			},
			expectRefund: true,
			expectError:  false,
		},
		{
			name: "should not fetch refund detail when refundId does not exist in metadata",
			setupMock: func() {
				metadataWithoutRefund := map[string]any{
					"other": "data",
				}

				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "payment-id", "merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "payment-id",
						MerchantID: "merchant-id",
						CustomerID: "customer-id",
						Amount:     decimal.NewFromFloat(100000),
						Currency:   "IDR",
						Status:     constant.StatusSuccess,
						CreatedAt:  now,
						UpdatedAt:  now,
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						},
						Metadata: &metadataWithoutRefund,
					}, nil).Once()

				mockPaymentRepo.On("GetChargeList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{},
				}, nil).Once()

				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "payment-id", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:                 uuid.New(),
						AccountID:            uuid.New(),
						ReferenceID:          "payment-id",
						TransactionTimestamp: now,
						Credit:               100000,
						Currency:             "IDR",
						Status:               constant.StatusSuccess,
					}, nil).Once()

				// mock status history
				mockStatusHistoriesRepo.On("GetByReference", mock.Anything, constant.TypePayment, "payment-id").
					Return([]*statusHistoryModel.StatusHistory{}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "merchant-id").Return(&customerModel.Customer{
					UUID: "customer-id",
				}, nil).Once()

				refundSvc.On("GetExistingRefundList", mock.Anything, mock.Anything).Return([]refundModel.RefundResponse{}, nil).Once()
			},
			expectRefund: false,
			expectError:  false,
		},
		{
			name: "should handle refund service error gracefully",
			setupMock: func() {
				refundId := "test-refund-id"
				metadataWithRefund := map[string]any{
					"refundId": refundId,
				}

				mockPaymentRepo.On("GetPaymentByIdAndMerchantId", mock.Anything, "payment-id", "merchant-id").
					Return(&paymentModel.Payment{
						UUID:       "payment-id",
						MerchantID: "merchant-id",
						CustomerID: "customer-id",
						Amount:     decimal.NewFromFloat(100000),
						Currency:   "IDR",
						Status:     constant.StatusSuccess,
						CreatedAt:  now,
						UpdatedAt:  now,
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						},
						Metadata: &metadataWithRefund,
					}, nil).Once()

				mockPaymentRepo.On("GetChargeList", mock.Anything, mock.Anything).Return(&commonModel.PaginationResponse{
					Data: []*unifiedPaymentModel.ChargeResponse{},
				}, nil).Once()

				mockAccountTransactionRepo.On("FindByReference", mock.Anything, "payment-id", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						UUID:                 uuid.New(),
						AccountID:            uuid.New(),
						ReferenceID:          "payment-id",
						TransactionTimestamp: now,
						Credit:               100000,
						Currency:             "IDR",
						Status:               constant.StatusSuccess,
					}, nil).Once()

				mockCustomerRepo.On("GetCustomerById", mock.Anything, "customer-id", "merchant-id").Return(&customerModel.Customer{
					UUID: "customer-id",
				}, nil).Once()

				refundSvc.On("GetExistingRefundList", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectRefund: false, // Should be nil when service fails
			expectError:  true,  // Should not fail the overall request
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			paymentSvc := PaymentService{
				logger:                 mockLogger,
				paymentRepo:            mockPaymentRepo,
				accountTransactionRepo: mockAccountTransactionRepo,
				customerRepo:           mockCustomerRepo,
				statusHistoriesRepo:    mockStatusHistoriesRepo,
				transferSvc:            transferSvc,
				refundSvc:              refundSvc,
			}

			ctx := context.Background()
			result, err := paymentSvc.GetPaymentHistoryDetail(ctx, paymentModel.PaymentHistoryDetailOption{
				PaymentID:  "payment-id",
				MerchantID: "merchant-id",
			})

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.expectRefund {
					assert.NotNil(t, result.RefundDetails, "Expected refund detail to be present")
					for _, refund := range result.RefundDetails {
						assert.NotEmpty(t, refund.ID)
						assert.NotEmpty(t, refund.ClientReferenceID)
						assert.NotEmpty(t, refund.PaymentSessionID)
					}
				} else {
					assert.Nil(t, result.RefundDetails, "Expected refund detail to be nil")
				}
			}

			// Assert all mocks were called
			mockPaymentRepo.AssertExpectations(t)
			mockAccountTransactionRepo.AssertExpectations(t)
			mockStatusHistoriesRepo.AssertExpectations(t)
			transferSvc.AssertExpectations(t)
			refundSvc.AssertExpectations(t)
		})
	}
}
