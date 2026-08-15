package orchestrator_model_test

import (
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"

	"github.com/stretchr/testify/assert"
)

func TestTransactionTypeForUser(t *testing.T) {
	tests := []struct {
		name    string
		typ     string
		channel string
		want    string
	}{
		{
			name: "top up",
			typ:  c.TypeTopUp,
			want: "VA Top Up",
		},
		{
			name: "disbursement",
			typ:  "DISBURSEMENT",
			want: "Single Payout",
		},
		{
			name: "bulk disbursement",
			typ:  "BULK_DISBURSEMENT",
			want: "Bulk Payout",
		},
		{
			name:    "manual top up",
			typ:     c.TypeManualAdjust,
			channel: c.ChannelManualTransfer,
			want:    "Manual Top Up",
		},
		{
			name:    "balance adjustment",
			typ:     c.TypeManualAdjust,
			channel: c.ChannelBalanceAdjustment,
			want:    "Balance Adjustment",
		},
		{
			name: "cross border fee",
			typ:  c.TypeXB + "_" + c.TypeFee,
			want: "Cross Border Fee",
		},
		{
			name:    "VA payment",
			typ:     c.TypePayment,
			channel: c.ChannelVirtualAccount,
			want:    "VA Payment",
		},
		{
			name:    "QRIS payment",
			typ:     c.TypePayment,
			channel: c.ChannelQris,
			want:    "QRIS Payment",
		},
		{
			name:    "credit card payment",
			typ:     c.TypePayment,
			channel: c.ChannelCreditCard,
			want:    "Cards Payment",
		},
		{
			name:    "payment with empty channel",
			typ:     c.TypePayment,
			channel: "",
			want:    "Payment",
		},
		{
			name:    "payment with unknown channel",
			typ:     c.TypePayment,
			channel: "SOME_UNKNOWN_CHANNEL",
			want:    "Payment",
		},
		{
			name: "unknown type with title case",
			typ:  "some_type",
			want: "Some Type",
		},
		{
			name: "unknown type with multiple words",
			typ:  "some_other_type",
			want: "Some Other Type",
		},
		{
			name: "Payout Fee Transactions",
			typ:  "DISBURSEMENT_FEE",
			want: "Payout Fee",
		},
		{
			name: "Refund",
			typ:  "REFUND",
			want: "Payment Refund",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransactionTypeForUser(tt.typ, tt.channel)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatChannelName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "QRIS special case",
			input: "QRIS",
			want:  "QRIS",
		},
		{
			name:  "XB special case",
			input: "XB",
			want:  "XB",
		},
		{
			name:  "PPOB special case",
			input: "PPOB",
			want:  "PPOB",
		},
		{
			name:  "virtual account",
			input: "VIRTUAL_ACCOUNT",
			want:  "Virtual Account",
		},
		{
			name:  "virtual account with spaces",
			input: "virtual_account",
			want:  "Virtual Account",
		},
		{
			name:  "virtual account with mixed case",
			input: "Virtual_Account",
			want:  "Virtual Account",
		},
		{
			name:  "credit card",
			input: "CREDIT_CARD",
			want:  "Credit Card",
		},
		{
			name:  "multiple words",
			input: "BANK_TRANSFER",
			want:  "Bank Transfer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatChannelName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
