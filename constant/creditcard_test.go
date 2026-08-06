package constant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChannelTypeToMidType(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SUCCESS: convert aggregator channel type to MID type",
			input:    PaymentMethodChannelTypeAggregator,
			expected: CreditCardMidTypeAggregator,
		},
		{
			name:     "SUCCESS: convert facilitator channel type to MID type",
			input:    PaymentMethodChannelTypeDirect,
			expected: CreditCardMidTypeDirect,
		},
		{
			name:     "SUCCESS: return input for unknown channel type",
			input:    "UNKNOWN_CHANNEL_TYPE",
			expected: "UNKNOWN_CHANNEL_TYPE",
		},
		{
			name:     "SUCCESS: return input for empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ChannelTypeToMidType(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMidTypeToChannelType(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SUCCESS: convert aggregator MID type to channel type",
			input:    CreditCardMidTypeAggregator,
			expected: PaymentMethodChannelTypeAggregator,
		},
		{
			name:     "SUCCESS: convert direct MID type to channel type",
			input:    CreditCardMidTypeDirect,
			expected: PaymentMethodChannelTypeDirect,
		},
		{
			name:     "SUCCESS: return input for unknown MID type",
			input:    "UNKNOWN_MID_TYPE",
			expected: "UNKNOWN_MID_TYPE",
		},
		{
			name:     "SUCCESS: return input for empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MidTypeToChannelType(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
