package fdscommon

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestCRMUpdateRequestToUpdateRequest(t *testing.T) {
	tests := []struct {
		name     string
		input    *CRMUpdateRequest
		expected *UpdateRequest
	}{
		{
			name: "with payment and all payment fields populated",
			input: &CRMUpdateRequest{
				AgentCode: "agent-001",
				IsFraud:   true,
				Status:    "approved",
				FraudType: "payment risk",
				Note:      "confirmed fraud",
				Payment: &CRMPaymentUpdate{
					CardStatus:       "stolen",
					PaymentStatus:    "chargeback",
					ChargebackStatus: "opened",
				},
			},
			expected: &UpdateRequest{
				AgentCode: util.ValueToPtr("agent-001"),
				IsFraud:   util.ValueToPtr(true),
				Status:    "approved",
				FraudType: util.ValueToPtr("payment risk"),
				Note:      util.ValueToPtr("confirmed fraud"),
				Payment: &PaymentUpdate{
					CardStatus:       util.ValueToPtr("stolen"),
					PaymentStatus:    "chargeback",
					ChargebackStatus: util.ValueToPtr("opened"),
				},
			},
		},
		{
			name: "with payment but empty card and chargeback status",
			input: &CRMUpdateRequest{
				AgentCode: "agent-002",
				IsFraud:   false,
				Status:    "cancelled",
				FraudType: "",
				Note:      "not fraud",
				Payment: &CRMPaymentUpdate{
					CardStatus:       "",
					PaymentStatus:    "paid",
					ChargebackStatus: "",
				},
			},
			expected: &UpdateRequest{
				AgentCode: util.ValueToPtr("agent-002"),
				IsFraud:   util.ValueToPtr(false),
				Status:    "cancelled",
				FraudType: util.ValueToPtr(""),
				Note:      util.ValueToPtr("not fraud"),
				Payment: &PaymentUpdate{
					CardStatus:       nil,
					PaymentStatus:    "paid",
					ChargebackStatus: nil,
				},
			},
		},
		{
			name: "without payment",
			input: &CRMUpdateRequest{
				AgentCode: "agent-003",
				IsFraud:   false,
				Status:    "new",
				FraudType: "identity theft",
				Note:      "review needed",
				Payment:   nil,
			},
			expected: &UpdateRequest{
				AgentCode: util.ValueToPtr("agent-003"),
				IsFraud:   util.ValueToPtr(false),
				Status:    "new",
				FraudType: util.ValueToPtr("identity theft"),
				Note:      util.ValueToPtr("review needed"),
				Payment:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToUpdateRequest()
			assert.Equal(t, tt.expected, result)
		})
	}
}
