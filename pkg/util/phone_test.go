package util

import "testing"

func TestCleanUpIDNPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "should remove +62 prefix",
			input:    "+62812345678",
			expected: "812345678",
		},
		{
			name:     "should remove 62 prefix",
			input:    "62812345678",
			expected: "812345678",
		},
		{
			name:     "should remove 0 prefix",
			input:    "0812345678",
			expected: "812345678",
		},
		{
			name:     "should not modify number without prefix",
			input:    "812345678",
			expected: "812345678",
		},
		{
			name:     "should handle empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "should handle single digit",
			input:    "5",
			expected: "5",
		},
		{
			name:     "should only remove prefix once",
			input:    "620812345678",
			expected: "0812345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanUpIDNPhoneNumber(tt.input)
			if result != tt.expected {
				t.Errorf("CleanUpIDNPhoneNumber(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
