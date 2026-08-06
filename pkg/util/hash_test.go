package util

import (
	"bytes"
	"io"
	"testing"
)

func TestIOToSHA256(t *testing.T) {
	tests := []struct {
		name     string
		input    io.Reader
		expected string
	}{
		{
			name:     "NilReader",
			input:    nil,
			expected: "",
		},
		{
			name:     "EmptyReader",
			input:    bytes.NewReader([]byte{}),
			expected: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "NonEmptyReader",
			input:    bytes.NewReader([]byte("hello world")),
			expected: "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IOToSHA256(tt.input)
			if result != tt.expected {
				t.Errorf("IOToSHA256() = %v, want %v", result, tt.expected)
			}
		})
	}
}
