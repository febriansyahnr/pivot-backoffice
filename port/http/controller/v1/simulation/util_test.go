package simulationController

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
)

func TestGetRedirectionURL(t *testing.T) {
	paymentSvc := serviceMocks.NewIPaymentService(t)
	handler := New(validatorExt.New(), WithPaymentService(paymentSvc))

	testCases := []struct {
		name         string
		paymentID    string
		chargeStatus string
		setupMock    func()
		expectedURL  string
		expectedErr  error
	}{
		{
			name:      "ERROR: GetDetailByID returns error",
			paymentID: "payment-id-1",
			setupMock: func() {
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), "payment-id-1").
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedURL: "",
			expectedErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name:      "ERROR: Payment not found",
			paymentID: "payment-id-2",
			setupMock: func() {
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), "payment-id-2").
					Once().Return(nil, nil)
			},
			expectedURL: "",
			expectedErr: constant.ErrPaymentNotFound,
		},
		{
			name:      "SUCCESS: Payment mode is not API - returns empty string",
			paymentID: "payment-id-3",
			setupMock: func() {
				metadata := map[string]any{
					"mode": "REDIRECT",
				}
				payment := &paymentModel.Payment{
					UUID:     "payment-id-3",
					Status:   constant.UnifiedPaymentSessionStatusPaid,
					Metadata: &metadata,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), "payment-id-3").
					Once().Return(payment, nil)
			},
			expectedURL: "",
			expectedErr: nil,
		},
		{
			name:      "ERROR: Invalid clientRedirectUrl metadata",
			paymentID: "payment-id-4",
			setupMock: func() {
				metadata := map[string]any{
					"mode":              constant.UnifiedPaymentModeAPI,
					"clientRedirectUrl": "invalid-structure",
				}
				payment := &paymentModel.Payment{
					UUID:     "payment-id-4",
					Status:   constant.UnifiedPaymentSessionStatusPaid,
					Metadata: &metadata,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), "payment-id-4").
					Once().Return(payment, nil)
			},
			expectedURL: "",
			expectedErr: errors.New("invalid character 'i' looking for beginning of value"),
		},
		{
			name:         "SUCCESS: Payment status is PAID - returns success URL",
			paymentID:    "payment-id-5",
			chargeStatus: "SUCCESS",
			setupMock: func() {
				metadata := map[string]any{
					"mode": constant.UnifiedPaymentModeAPI,
					"clientRedirectUrl": map[string]any{
						"successReturnUrl":    "https://example.com/success",
						"failureReturnUrl":    "https://example.com/failure",
						"expirationReturnUrl": "https://example.com/expired",
					},
				}
				payment := &paymentModel.Payment{
					UUID:     "payment-id-5",
					Status:   constant.UnifiedPaymentSessionStatusPaid,
					Metadata: &metadata,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), "payment-id-5").
					Once().Return(payment, nil)
			},
			expectedURL: "https://example.com/success",
			expectedErr: nil,
		},
		{
			name:         "SUCCESS: Payment status is CANCELLED - returns failure URL",
			paymentID:    "payment-id-6",
			chargeStatus: "FAILED",
			setupMock: func() {
				metadata := map[string]any{
					"mode": constant.UnifiedPaymentModeAPI,
					"clientRedirectUrl": map[string]any{
						"successReturnUrl":    "https://example.com/success",
						"failureReturnUrl":    "https://example.com/failure",
						"expirationReturnUrl": "https://example.com/expired",
					},
				}
				payment := &paymentModel.Payment{
					UUID:     "payment-id-6",
					Status:   constant.UnifiedPaymentSessionStatusCancelled,
					Metadata: &metadata,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), "payment-id-6").
					Once().Return(payment, nil)
			},
			expectedURL: "https://example.com/failure",
			expectedErr: nil,
		},
		{
			name:         "SUCCESS: Payment status is EXPIRED - returns expiration URL",
			paymentID:    "payment-id-7",
			chargeStatus: "EXPIRED",
			setupMock: func() {
				metadata := map[string]any{
					"mode": constant.UnifiedPaymentModeAPI,
					"clientRedirectUrl": map[string]any{
						"successReturnUrl":    "https://example.com/success",
						"failureReturnUrl":    "https://example.com/failure",
						"expirationReturnUrl": "https://example.com/expired",
					},
				}
				payment := &paymentModel.Payment{
					UUID:     "payment-id-7",
					Status:   constant.UnifiedPaymentSessionStatusExpired,
					Metadata: &metadata,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), "payment-id-7").
					Once().Return(payment, nil)
			},
			expectedURL: "https://example.com/expired",
			expectedErr: nil,
		},
		{
			name:         "SUCCESS: Payment status is other - returns empty string",
			paymentID:    "payment-id-8",
			chargeStatus: "OTHER",
			setupMock: func() {
				metadata := map[string]any{
					"mode": constant.UnifiedPaymentModeAPI,
					"clientRedirectUrl": map[string]any{
						"successReturnUrl":    "https://example.com/success",
						"failureReturnUrl":    "https://example.com/failure",
						"expirationReturnUrl": "https://example.com/expired",
					},
				}
				payment := &paymentModel.Payment{
					UUID:     "payment-id-8",
					Status:   "PENDING",
					Metadata: &metadata,
				}
				paymentSvc.On("GetDetailByID", constant.ValueCtxMockType(), "payment-id-8").
					Once().Return(payment, nil)
			},
			expectedURL: "",
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setupMock()

			url, err := handler.(*Handler).GetRedirectionURL(ctx, tc.paymentID, tc.chargeStatus)

			assert.Equal(t, tc.expectedURL, url)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			paymentSvc.AssertExpectations(t)
		})
	}
}
