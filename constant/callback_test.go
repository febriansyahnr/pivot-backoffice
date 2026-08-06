package constant_test

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCallbackEventTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: constant.CallbackEventPayoutDone,
			want:  "Payout Done",
		},
		{
			input: constant.CallbackEventPayoutPending,
			want:  "Payout Pending",
		},
		{
			input: constant.CallbackEventPaymentVirtualAccountPaid,
			want:  "Payment Virtual Account Paid",
		},
		{
			input: "OTHER",
			want:  "",
		},
	}
	for _, test := range tests {
		eventTitle := constant.GetCallbackEventTitle(test.input)
		assert.Equal(t, test.want, eventTitle)
	}
}
