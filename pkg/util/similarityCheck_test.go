package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimilarityCheck(t *testing.T) {
	tests := []struct {
		name       string
		inputName  string
		bankRecord string
		expected   string
	}{
		{
			name:       "Exact match",
			inputName:  "John Doe",
			bankRecord: "John Doe",
			expected:   "",
		},
		{
			name:       "Similar names with prefix",
			inputName:  "Sdr. John Doe",
			bankRecord: "John Doe",
			expected:   "",
		},
		{
			name:       "Completely different names",
			inputName:  "Alice Smith",
			bankRecord: "Bob Johnson",
			expected:   "The account name entered does not match the bank's records. Please check the account information and try again. Bank record: Bob Johnson",
		},
		{
			name:       "Similar but not exact match",
			inputName:  "John Does",
			bankRecord: "John Doe",
			expected:   "The account name entered is not an exact match. Please check the account information and try again. Bank record: John Doe",
		},
		{
			name:       "Empty strings",
			inputName:  "",
			bankRecord: "",
			expected:   "The account name entered is an exact match. Bank record: ",
		},
		{
			name:       "Empty input name",
			inputName:  "",
			bankRecord: "John Doe",
			expected:   "The account name entered does not match the bank's records. Please check the account information and try again. Bank record: John Doe",
		},
		{
			name:       "Empty bank record",
			inputName:  "John Doe",
			bankRecord: "",
			expected:   "The account name entered does not match the bank's records. Please check the account information and try again. Bank record: ",
		},
		{
			name:       "Names with different cases",
			inputName:  "JOHN DOE",
			bankRecord: "john doe",
			expected:   "",
		},
		{
			name:       "Same name with prefix case insensitive",
			inputName:  "SDR FEBRIANI DWI SAFITRI",
			bankRecord: "FEBRIANI DWI SAFITRI",
			expected:   "",
		},
		{
			name:       "Same name with prefix case insensitive 2",
			inputName:  "SUSANTI",
			bankRecord: "Ibu SUSANTI",
			expected:   "",
		},
		{
			name:       "Same name with prefix case insensitive 3",
			inputName:  "Ibu. SUSANTI",
			bankRecord: "IBU SUSANTI",
			expected:   "",
		},
		{
			name:       "Same name, extra space from bank name",
			inputName:  "SDR FEBRIANI  DWI  SAFITRI",
			bankRecord: "FEBRIANI DWI  SAFITRI",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, result := SimilarityCheck(tt.inputName, tt.bankRecord, "", "")
			assert.Equalf(t, tt.expected, result, "SimilarityCheck(%q, %q) output is %q, message is [%q], want %q", tt.inputName, tt.bankRecord, status, result, tt.expected)
		})
	}
}

func TestCheckPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefixes []string
		str      string
		expected bool
	}{
		{
			name:     "Success",
			str:      "sdr. John",
			expected: true,
		},
		{
			name:     "Failed match in middle",
			str:      "Iwan sdrian",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NameHasPrefix(tt.str)
			if result != tt.expected {
				t.Errorf("NameHasPrefix(%v, %q) = %v, want %v", tt.prefixes, tt.str, result, tt.expected)
			}
		})
	}
}
