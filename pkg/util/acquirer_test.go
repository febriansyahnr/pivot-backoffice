package util

import "testing"

func TestCleanAcquirerName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "BNC with QRIS suffix",
			input:    "BNC_QRIS",
			expected: "bnc",
		},
		{
			name:     "BNI with QRIS suffix",
			input:    "BNI_QRIS",
			expected: "bni",
		},
		{
			name:     "BRI with VA suffix",
			input:    "BRI_VA",
			expected: "bri",
		},
		{
			name:     "BNC with EWALLET suffix",
			input:    "BNC_EWALLET",
			expected: "bnc",
		},
		{
			name:     "BNI with CC suffix",
			input:    "BNI_CC",
			expected: "bni",
		},
		{
			name:     "OVO with OVO suffix",
			input:    "OVO_OVO",
			expected: "ovo",
		},
		{
			name:     "DANA with DANA suffix",
			input:    "DANA_DANA",
			expected: "dana",
		},
		{
			name:     "GOPAY with GOPAY suffix",
			input:    "GOPAY_GOPAY",
			expected: "gopay",
		},
		{
			name:     "LINKAJA with LINKAJA suffix",
			input:    "LINKAJA_LINKAJA",
			expected: "linkaja",
		},
		{
			name:     "Plain BNC without suffix",
			input:    "BNC",
			expected: "bnc",
		},
		{
			name:     "Plain BNI without suffix",
			input:    "BNI",
			expected: "bni",
		},
		{
			name:     "Lowercase input",
			input:    "bnc_qris",
			expected: "bnc",
		},
		{
			name:     "Mixed case input",
			input:    "Bnc_Qris",
			expected: "bnc",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanAcquirerName(tt.input)
			if result != tt.expected {
				t.Errorf("CleanAcquirerName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
