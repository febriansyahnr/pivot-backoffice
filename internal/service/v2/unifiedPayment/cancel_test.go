package unifiedPaymentService_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	unifiedPaymentService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelSession(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	merchantID := "merchant-" + uuid.NewString()
	paymentID := "payment-" + uuid.NewString()
	cancellationReason := "REQUESTED_BY_CUSTOMER"
	databaseError := errors.New("database error")
	now := time.Now()

	type Mockers struct {
		paymentRepo       *repositoryMock.IPaymentRepository
		paymentMethodRepo *repositoryMock.IPaymentMethodRepository
		accountTrxRepo    *repositoryMock.IAccountTransactionRepository
	}

	testCases := []struct {
		desc             string
		request          *unifiedPaymentModel.CancelUnifiedPaymentSessionRequest
		wantError        bool
		wantNilResult    bool
		expectedErrType  string
		expectedErrMsg   string
		setupMock        func(mockers Mockers)
	}{
		{
			desc: "database error when getting payment",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:       true,
			wantNilResult:   true,
			expectedErrType: response.HttpErrDatabase,
			expectedErrMsg:  databaseError.Error(),
			setupMock: func(mockers Mockers) {
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(nil, databaseError)
			},
		},
		{
			desc: "payment not found should return unprocessable",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:       true,
			wantNilResult:   true,
			expectedErrType: response.HttpErrUnprocessableContent,
			expectedErrMsg:  c.ErrPaymentNotFound.Error(),
			setupMock: func(mockers Mockers) {
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(nil, nil)
			},
		},
		{
			desc: "merchant mismatch should return unprocessable",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:       true,
			wantNilResult:   true,
			expectedErrType: response.HttpErrUnprocessableContent,
			expectedErrMsg:  c.ErrMerchantIsNotMatch.Error(),
			setupMock: func(mockers Mockers) {
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: "different-merchant",
				}, nil)
			},
		},
		{
			desc: "REQUIRE_ACTION with API mode should return error",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:       true,
			wantNilResult:   true,
			expectedErrType: response.HttpErrRequest,
			expectedErrMsg:  "Payment session in REQUIRE_ACTION status can only be cancelled by customer on payment page when using redirection method",
			setupMock: func(mockers Mockers) {
				metadata := map[string]interface{}{"mode": c.UnifiedPaymentModeAPI}
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					Metadata:   &metadata,
				}, nil)
			},
		},
		{
			desc: "PROCESSING status cannot be cancelled",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:       true,
			wantNilResult:   true,
			expectedErrType: response.HttpErrUnprocessableContent,
			expectedErrMsg:  "Payment session in PROCESSING status cannot be cancelled",
			setupMock: func(mockers Mockers) {
				metadata := map[string]interface{}{"mode": c.UnifiedPaymentModeAPI}
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusProcessing,
					Metadata:   &metadata,
				}, nil)
			},
		},
		{
			desc: "PAID status cannot be cancelled",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:       true,
			wantNilResult:   true,
			expectedErrType: response.HttpErrUnprocessableContent,
			expectedErrMsg:  "Payment session in PAID status cannot be cancelled",
			setupMock: func(mockers Mockers) {
				metadata := map[string]interface{}{"mode": c.UnifiedPaymentModeAPI}
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusPaid,
					Metadata:   &metadata,
				}, nil)
			},
		},
		{
			desc: "CANCELLED status cannot be cancelled",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:       true,
			wantNilResult:   true,
			expectedErrType: response.HttpErrUnprocessableContent,
			expectedErrMsg:  "Payment session in CANCELLED status cannot be cancelled",
			setupMock: func(mockers Mockers) {
				metadata := map[string]interface{}{"mode": c.UnifiedPaymentModeAPI}
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusCancelled,
					Metadata:   &metadata,
				}, nil)
			},
		},
		{
			desc: "EXPIRED status cannot be cancelled",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:       true,
			wantNilResult:   true,
			expectedErrType: response.HttpErrUnprocessableContent,
			expectedErrMsg:  "Payment session in EXPIRED status cannot be cancelled",
			setupMock: func(mockers Mockers) {
				metadata := map[string]interface{}{"mode": c.UnifiedPaymentModeAPI}
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusExpired,
					Metadata:   &metadata,
				}, nil)
			},
		},
		{
			desc: "database error when updating payment status",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:       true,
			wantNilResult:   true,
			expectedErrType: response.HttpErrDatabase,
			expectedErrMsg:  databaseError.Error(),
			setupMock: func(mockers Mockers) {
				metadata := map[string]interface{}{"mode": c.UnifiedPaymentModeAPI}
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequirePaymentMethod,
					Metadata:   &metadata,
					CreatedAt:  now,
					UpdatedAt:  now,
				}, nil)
				mockers.paymentRepo.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(databaseError)
			},
		},
		{
			desc: "success with REQUIRE_PAYMENT_METHOD status",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:     false,
			wantNilResult: false,
			setupMock: func(mockers Mockers) {
				metadata := map[string]interface{}{"mode": c.UnifiedPaymentModeAPI}
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequirePaymentMethod,
					Metadata:   &metadata,
					CreatedAt:  now,
					UpdatedAt:  now,
				}, nil)
				mockers.paymentRepo.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
			},
		},
		{
			desc: "success with REQUIRE_CONFIRMATION status",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "MERCHANT",
			},
			wantError:     false,
			wantNilResult: false,
			setupMock: func(mockers Mockers) {
				metadata := map[string]interface{}{"mode": c.UnifiedPaymentModeAPI}
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireConfirmation,
					Metadata:   &metadata,
					CreatedAt:  now,
					UpdatedAt:  now,
				}, nil)
				mockers.paymentRepo.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
			},
		},
		{
			desc: "success with REQUIRE_ACTION redirect mode and customer source",
			request: &unifiedPaymentModel.CancelUnifiedPaymentSessionRequest{
				PaymentSessionID:   paymentID,
				MerchantID:         merchantID,
				CancellationReason: cancellationReason,
				Source:             "CUSTOMER",
			},
			wantError:     false,
			wantNilResult: false,
			setupMock: func(mockers Mockers) {
				metadata := map[string]interface{}{"mode": c.UnifiedPaymentModeRedirect}
				mockers.paymentRepo.On("GetPaymentById", mock.Anything, paymentID).Return(&paymentModel.Payment{
					UUID:       paymentID,
					MerchantID: merchantID,
					Status:     c.UnifiedPaymentSessionStatusRequireAction,
					Metadata:   &metadata,
					CreatedAt:  now,
					UpdatedAt:  now,
				}, nil)
				mockers.paymentRepo.On("UpdatePaymentData", mock.Anything, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockers := Mockers{
				paymentRepo:       repositoryMock.NewIPaymentRepository(t),
				paymentMethodRepo: repositoryMock.NewIPaymentMethodRepository(t),
				accountTrxRepo:    repositoryMock.NewIAccountTransactionRepository(t),
			}
			tc.setupMock(mockers)

			svc := unifiedPaymentService.New(cfg, log, mockers.paymentRepo, mockers.paymentMethodRepo, mockers.accountTrxRepo)
			result, err := svc.CancelSession(context.Background(), tc.request)

			if tc.wantError {
				assert.Error(t, err)
				errType, rawErr := pkgErrors.ExtractError(err)
				assert.Equal(t, tc.expectedErrType, errType)
				assert.Equal(t, tc.expectedErrMsg, rawErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tc.wantNilResult {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}

			mockers.paymentRepo.AssertExpectations(t)
		})
	}
}
