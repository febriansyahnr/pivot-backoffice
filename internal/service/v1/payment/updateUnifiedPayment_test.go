package paymentService

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/test"
)

func TestAllowedToUpdateMethodOrPaymentOptionExtraScenarios(t *testing.T) {
	// Setup
	_, pdkLog, _ := test.SetupLogger()
	ctx := context.Background()

	// Create metadata map
	snapCoreMetadata := map[string]interface{}{
		"snapCore": map[string]any{
			"acquirer": "permata",
		},
	}

	testCases := []struct {
		name            string
		request         *paymentModel.UpdateUnifiedPaymentRequest
		existingPayment *paymentModel.Payment
		expected        bool
	}{
		{
			name: "Same payment method but different VA issuer",
			request: &paymentModel.UpdateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "bca",
					},
				},
			},
			existingPayment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Metadata: &snapCoreMetadata,
			},
			expected: true,
		},
		{
			name: "Same payment method and same VA issuer",
			request: &paymentModel.UpdateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "permata",
					},
				},
			},
			existingPayment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Metadata: &snapCoreMetadata,
			},
			expected: false,
		},
		{
			name: "Same payment method with case insensitive VA issuer",
			request: &paymentModel.UpdateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "PERMATA",
					},
				},
			},
			existingPayment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Metadata: &snapCoreMetadata,
			},
			expected: false,
		},
		{
			name: "Credit card payment method should allow update",
			request: &paymentModel.UpdateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			},
			existingPayment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			expected: true,
		},
		{
			name: "QRIS payment method should allow update",
			request: &paymentModel.UpdateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
			},
			existingPayment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_QRIS,
				},
			},
			expected: true,
		},
		{
			name: "VA payment with empty metadata",
			request: &paymentModel.UpdateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "bri",
					},
				},
			},
			existingPayment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Metadata: &map[string]interface{}{},
			},
			expected: false,
		},
		{
			name: "VA payment with missing snapCore in metadata",
			request: &paymentModel.UpdateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "bni",
					},
				},
			},
			existingPayment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Metadata: &map[string]interface{}{
					"otherData": "value",
				},
			},
			expected: false,
		},
		{
			name: "Same VA payment method with invalid metadata",
			request: &paymentModel.UpdateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{
						Issuer: "bca",
					},
				},
			},
			existingPayment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Metadata: &map[string]interface{}{
					"snapCore": []byte(`invalid json`),
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create service with minimal setup
			service := &PaymentService{
				logger: pdkLog,
			}

			result := service.AllowedToUpdateMethodOrPaymentOption(ctx, tc.request, tc.existingPayment)

			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestIsUpdateEligible(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	testCases := []struct {
		name     string
		payment  *paymentModel.Payment
		expected bool
	}{
		{
			name: "Credit card payment with pending status",
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
				Status:    paymentConstant.PAYMENT_STATUS_PENDING,
				ExpiredAt: &future,
			},
			expected: true,
		},
		{
			name: "Credit card payment with non-pending status",
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
				Status:    paymentConstant.PAYMENT_STATUS_SUCCESS,
				ExpiredAt: &future,
			},
			expected: false,
		},
		{
			name: "VA payment with waiting for payment status",
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Status:    paymentConstant.UnifiedPaymentStatusWaitingForPayment,
				ExpiredAt: &future,
			},
			expected: true,
		},
		{
			name: "VA payment with non-waiting status",
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Status:    paymentConstant.PAYMENT_STATUS_SUCCESS,
				ExpiredAt: &future,
			},
			expected: false,
		},
		{
			name: "QRIS payment with waiting for payment status",
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_QRIS,
				},
				Status:    paymentConstant.UnifiedPaymentStatusWaitingForPayment,
				ExpiredAt: &future,
			},
			expected: true,
		},
		{
			name: "Expired payment",
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Status:    paymentConstant.UnifiedPaymentStatusWaitingForPayment,
				ExpiredAt: &past,
			},
			expected: false,
		},
		{
			name: "Expired credit card payment",
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
				Status:    paymentConstant.PAYMENT_STATUS_PENDING,
				ExpiredAt: &past,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create service with minimal setup
			service := &PaymentService{}

			result := service.IsUpdateEligible(tc.payment)

			assert.Equal(t, tc.expected, result)
		})
	}
}
