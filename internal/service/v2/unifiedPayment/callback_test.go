//nolint:testpackage // test accesses unexported methods isAutoSplitPaymentAllowedSendCallback, shouldSkipSendCallback
package unifiedPaymentService

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"

	"github.com/stretchr/testify/assert"
)

func TestIsAutoSplitPaymentAllowedSendCallback(t *testing.T) {
	svc := UnifiedPaymentService{}

	tests := []struct {
		name    string
		payment *paymentModel.Payment
		want    bool
	}{
		{
			name: "returns false when AutoSplitPayment is nil",
			payment: &paymentModel.Payment{
				AutoSplitPayment: nil,
			},
			want: false,
		},
		{
			name: "returns false when payment is auto-split sub-payment (FIRST_PAYMENT)",
			payment: &paymentModel.Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeFirstPayment,
				},
			},
			want: false,
		},
		{
			name: "returns false when payment is auto-split sub-payment (SUBSEQUENCE_PAYMENT)",
			payment: &paymentModel.Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeSubsequentPayment,
				},
			},
			want: false,
		},
		{
			name: "returns false when payment is not auto-split auth",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusPaid,
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: "OTHER_TYPE",
					Summary:         &unifiedPaymentModel.AutoSplitPaymentSummary{Status: constant.AutoSplitPaymentStatusSuccess},
				},
			},
			want: false,
		},
		{
			name: "returns true when auth + payment status is PROCESSING",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
					Summary:         &unifiedPaymentModel.AutoSplitPaymentSummary{Status: constant.AutoSplitPaymentStatusProcessing},
				},
			},
			want: true,
		},
		{
			name: "returns true when auth + summary status is not PROCESSING",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusPaid,
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
					Summary:         &unifiedPaymentModel.AutoSplitPaymentSummary{Status: constant.AutoSplitPaymentStatusSuccess},
				},
			},
			want: true,
		},
		{
			name: "returns false when auth + non-PROCESSING payment + summary still PROCESSING",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusPaid,
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
					Summary:         &unifiedPaymentModel.AutoSplitPaymentSummary{Status: constant.AutoSplitPaymentStatusProcessing},
				},
			},
			want: false,
		},
		{
			name: "returns true when auth + summary status is FAILED",
			payment: &paymentModel.Payment{
				Status: constant.UnifiedPaymentSessionStatusPaid,
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
					Summary:         &unifiedPaymentModel.AutoSplitPaymentSummary{Status: constant.AutoSplitPaymentStatusFailed},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.isAutoSplitPaymentAllowedSendCallback(tt.payment)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShouldSkipSendCallback(t *testing.T) {
	svc := UnifiedPaymentService{}

	tests := []struct {
		name    string
		payment *paymentModel.Payment
		want    bool
	}{
		{
			name: "skips callback for VIRTUAL_TERMINAL type",
			payment: &paymentModel.Payment{
				Type:             constant.TypeVirtualTerminal,
				AutoSplitPayment: nil,
			},
			want: true,
		},
		{
			name: "skips callback for CARD_FUNDED_PAYOUT type",
			payment: &paymentModel.Payment{
				Type:             constant.TypeCardFundedPayout,
				AutoSplitPayment: nil,
			},
			want: true,
		},
		{
			name: "skips callback for ONE_DOLLAR_AUTHORIZATION type",
			payment: &paymentModel.Payment{
				Type:             constant.UnifiedPaymentOneDollarAuthorization,
				AutoSplitPayment: nil,
			},
			want: true,
		},
		{
			name: "does not skip for regular payment type without auto-split",
			payment: &paymentModel.Payment{
				Type:             "PAYMENT",
				AutoSplitPayment: nil,
			},
			want: false,
		},
		{
			name: "does not skip for allowed auto-split payment",
			payment: &paymentModel.Payment{
				Type:   "PAYMENT",
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
					Summary:         &unifiedPaymentModel.AutoSplitPaymentSummary{Status: constant.AutoSplitPaymentStatusProcessing},
				},
			},
			want: false,
		},
		{
			name: "skips for auth + non-PROCESSING payment + summary still PROCESSING",
			payment: &paymentModel.Payment{
				Type:   "PAYMENT",
				Status: constant.UnifiedPaymentSessionStatusPaid,
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
					Summary:         &unifiedPaymentModel.AutoSplitPaymentSummary{Status: constant.AutoSplitPaymentStatusProcessing},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.shouldSkipSendCallback(tt.payment)
			assert.Equal(t, tt.want, got)
		})
	}
}
