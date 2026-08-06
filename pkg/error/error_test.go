package errors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		errType  string
		err      error
		expected string
	}{
		{
			name:     "should create new error with type",
			errType:  "TYPE_ERROR",
			err:      errors.New("some error"),
			expected: "TYPE_ERROR | some error",
		},
		{
			name:     "should handle empty error type",
			errType:  "",
			err:      errors.New("some error"),
			expected: " | some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.errType, tt.err)
			if result.Error() != tt.expected {
				t.Errorf("New() = %v, want %v", result.Error(), tt.expected)
			}
		})
	}
}

func TestExtractError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedType  string
		expectedError string
	}{
		{
			name:          "should extract error type and message",
			err:           errors.New("TYPE_ERROR | some error"),
			expectedType:  "TYPE_ERROR",
			expectedError: "some error",
		},
		{
			name:          "should handle error without type",
			err:           errors.New("some error"),
			expectedType:  "",
			expectedError: "some error",
		},
		{
			name:          "should handle multiple delimiters",
			err:           errors.New("TYPE_ERROR | error | with | multiple | parts"),
			expectedType:  "TYPE_ERROR",
			expectedError: "error | with | multiple | parts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errType, extractedErr := ExtractError(tt.err)
			if errType != tt.expectedType {
				t.Errorf("ExtractError() type = %v, want %v", errType, tt.expectedType)
			}
			if extractedErr.Error() != tt.expectedError {
				t.Errorf("ExtractError() error = %v, want %v", extractedErr.Error(), tt.expectedError)
			}
		})
	}
}
