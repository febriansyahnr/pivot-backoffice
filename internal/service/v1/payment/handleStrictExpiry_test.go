package paymentService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleStrictExpiry(t *testing.T) {
	paymentID := uuid.New().String()
	merchantID := uuid.New().String()
	ledgerID := uuid.New()

	tests := []struct {
		name          string
		paymentID     string
		mockSetup     func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository)
		expectedError error
	}{
		{
			name:      "SUCCESS: Handle strict expiry for regular payment",
			paymentID: paymentID,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					Metadata: &map[string]any{
						"expirationMode": constant.UnifiedPaymentExpirationModeStrict,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentID, merchantID, paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock status history recording
				mockStatusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction
				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:           ledgerID,
					MerchantID:     uuid.MustParse(merchantID),
					AdditionalInfo: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, paymentID, constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:      "SUCCESS: Handle strict expiry for static payment (ACTIVE -> INACTIVE)",
			paymentID: paymentID,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     constant.UnifiedStaticPaymentStatusActive,
					Metadata: &map[string]any{
						"expirationMode": constant.UnifiedPaymentExpirationModeStrict,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentID, merchantID, constant.UnifiedStaticPaymentStatusInactive, mock.Anything).Return(nil).Once()

				// Mock status history recording
				mockStatusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction
				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:           ledgerID,
					MerchantID:     uuid.MustParse(merchantID),
					AdditionalInfo: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, paymentID, constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:      "SUCCESS: Handle strict expiry when payment ledger not found",
			paymentID: paymentID,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					Metadata: &map[string]any{
						"expirationMode": constant.UnifiedPaymentExpirationModeStrict,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentID, merchantID, paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock status history recording
				mockStatusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction not found (returns nil, nil)
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, paymentID, constant.TypePayment).Return(nil, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:      "ERROR: Database error when getting payment",
			paymentID: paymentID,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(nil, errors.New("database error")).Once()
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, errors.New("database error")),
		},
		{
			name:      "ERROR: Payment not found",
			paymentID: paymentID,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				mockPaymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(nil, nil).Once()
			},
			expectedError: pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentNotFound),
		},
		{
			name:      "ERROR: Payment expiration mode is not strict",
			paymentID: paymentID,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					Metadata: &map[string]any{
						"expirationMode": "LOOSE", // Not strict mode
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil).Once()
			},
			expectedError: pkgErrors.New(response.HttpErrRequest, errors.New("payment expiration mode is not strict")),
		},
		{
			name:      "ERROR: Failed to update payment status",
			paymentID: paymentID,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					Metadata: &map[string]any{
						"expirationMode": constant.UnifiedPaymentExpirationModeStrict,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentID, merchantID, paymentConstant.PaymentStatusExpired, mock.Anything).Return(errors.New("update failed")).Once()
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, errors.New("update failed")),
		},
		{
			name:      "ERROR: Failed to find payment ledger by reference",
			paymentID: paymentID,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					Metadata: &map[string]any{
						"expirationMode": constant.UnifiedPaymentExpirationModeStrict,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentID, merchantID, paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock status history recording
				mockStatusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction error
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, paymentID, constant.TypePayment).Return(nil, errors.New("database error")).Once()
			},
			expectedError: pkgErrors.New(response.HttpStatusErrorNotFound, errors.New("database error")),
		},
		{
			name:      "ERROR: Failed to update payment transaction status and metadata",
			paymentID: paymentID,
			mockSetup: func(mockPaymentRepo *repositoryMocks.IPaymentRepository, mockAccountTransactionRepo *repositoryMocks.IAccountTransactionRepository, mockStatusHistoryRepo *repositoryMocks.IStatusHistoriesRepository) {
				payment := &paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     paymentConstant.PAYMENT_STATUS_PENDING,
					Metadata: &map[string]any{
						"expirationMode": constant.UnifiedPaymentExpirationModeStrict,
					},
				}
				mockPaymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(payment, nil).Once()
				mockPaymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentID, merchantID, paymentConstant.PaymentStatusExpired, mock.Anything).Return(nil).Once()

				// Mock status history recording
				mockStatusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

				// Mock account transaction
				accountTransaction := &orchestrator_model.AccountTransactionWithUseCase{
					UUID:           ledgerID,
					MerchantID:     uuid.MustParse(merchantID),
					AdditionalInfo: types.NullJSONText{Valid: true, JSONText: []byte(`{}`)},
				}
				mockAccountTransactionRepo.On("FindByReference", mock.Anything, paymentID, constant.TypePayment).Return(accountTransaction, nil).Once()
				mockAccountTransactionRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("update metadata failed")).Once()
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, errors.New("update metadata failed")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockPaymentRepo := repositoryMocks.NewIPaymentRepository(t)
			mockAccountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockStatusHistoryRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			service := &PaymentService{
				paymentRepo:            mockPaymentRepo,
				accountTransactionRepo: mockAccountTransactionRepo,
				statusHistoriesRepo:    mockStatusHistoryRepo,
				logger:                 mockLogger,
			}

			tc.mockSetup(mockPaymentRepo, mockAccountTransactionRepo, mockStatusHistoryRepo)
			err := service.HandleStrictExpiry(context.Background(), tc.paymentID)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockAccountTransactionRepo.AssertExpectations(t)
			mockStatusHistoryRepo.AssertExpectations(t)
		})
	}
}
