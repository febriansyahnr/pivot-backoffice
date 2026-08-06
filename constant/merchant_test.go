package constant_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/stretchr/testify/assert"
)

func TestIsValidRiskLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid risk level LOW",
			input:    constant.MerchantRiskLevelLow,
			expected: true,
		},
		{
			name:     "Valid risk level LOW_MID",
			input:    constant.MerchantRiskLevelLowMid,
			expected: true,
		},
		{
			name:     "Valid risk level MID",
			input:    constant.MerchantRiskLevelMid,
			expected: true,
		},
		{
			name:     "Valid risk level MID_HIGH",
			input:    constant.MerchantRiskLevelMidHigh,
			expected: true,
		},
		{
			name:     "Valid risk level HIGH",
			input:    constant.MerchantRiskLevelHigh,
			expected: true,
		},
		{
			name:     "Empty risk level should be allowed",
			input:    "",
			expected: true,
		},
		{
			name:     "Invalid risk level",
			input:    "INVALID_LEVEL",
			expected: false,
		},
		{
			name:     "Lowercase invalid risk level",
			input:    "low",
			expected: false,
		},
		{
			name:     "Mixed case invalid risk level",
			input:    "Low",
			expected: false,
		},
		{
			name:     "Random string",
			input:    "random_string",
			expected: false,
		},
		{
			name:     "Numeric string",
			input:    "123",
			expected: false,
		},
		{
			name:     "Special characters",
			input:    "HIGH!@#",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constant.IsValidRiskLevel(tt.input)
			assert.Equal(t, tt.expected, result, "IsValidRiskLevel(%q) = %v, want %v", tt.input, result, tt.expected)
		})
	}
}

func TestValidMerchantRiskLevels(t *testing.T) {
	expectedLevels := []string{
		constant.MerchantRiskLevelLow,
		constant.MerchantRiskLevelLowMid,
		constant.MerchantRiskLevelMid,
		constant.MerchantRiskLevelMidHigh,
		constant.MerchantRiskLevelHigh,
	}

	assert.Equal(t, expectedLevels, constant.ValidMerchantRiskLevels, "ValidMerchantRiskLevels should contain all expected risk levels")
	assert.Len(t, constant.ValidMerchantRiskLevels, 5, "ValidMerchantRiskLevels should contain exactly 5 levels")

	// Test that all levels in the slice are valid when passed to IsValidRiskLevel
	for _, level := range constant.ValidMerchantRiskLevels {
		assert.True(t, constant.IsValidRiskLevel(level), "Risk level %q should be valid", level)
	}
}