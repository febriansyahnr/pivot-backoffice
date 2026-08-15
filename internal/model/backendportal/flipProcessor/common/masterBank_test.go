package flipProcessorModel_test

import (
	"testing"

	flipProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/flipProcessor/common"
	"github.com/stretchr/testify/assert"
)

func TestGetFlipBankCode(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "success get flip code for BCA",
			input:    "014",
			expected: "bca",
		},
		{
			desc:     "success get flip code for BRI",
			input:    "002",
			expected: "bri",
		},
		{
			desc:     "not found",
			input:    "999",
			expected: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := flipProcessorModel.GetFlipBankCode(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestGetBankCode(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "success get bank code for BCA",
			input:    "bca",
			expected: "014",
		},
		{
			desc:     "success get bank code for BRI",
			input:    "bri",
			expected: "002",
		},
		{
			desc:     "not found",
			input:    "unknown",
			expected: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := flipProcessorModel.GetBankCode(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}
