package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatRupiah(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected string
	}{
		{
			name:     "small_amount",
			amount:   100,
			expected: "Rp 100,-",
		},
		{
			name:     "thousands",
			amount:   10000,
			expected: "Rp 10.000,-",
		},
		{
			name:     "hundreds_of_thousands",
			amount:   500000,
			expected: "Rp 500.000,-",
		},
		{
			name:     "millions",
			amount:   1000000,
			expected: "Rp 1.000.000,-",
		},
		{
			name:     "large_amount",
			amount:   999999999,
			expected: "Rp 999.999.999,-",
		},
		{
			name:     "zero_amount",
			amount:   0,
			expected: "Rp 0,-",
		},
		{
			name:     "decimal_truncated",
			amount:   10000.99,
			expected: "Rp 10.000,-",
		},
		{
			name:     "decimal_rounded_down",
			amount:   10000.49,
			expected: "Rp 10.000,-",
		},
		{
			name:     "single_digit",
			amount:   7,
			expected: "Rp 7,-",
		},
		{
			name:     "negative_amount",
			amount:   -5000,
			expected: "Rp -5.000,-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRupiah(tt.amount)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatRupiahWithoutDecimal(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected string
	}{
		{
			name:     "small_amount",
			amount:   100,
			expected: "Rp 100",
		},
		{
			name:     "thousands",
			amount:   10000,
			expected: "Rp 10.000",
		},
		{
			name:     "hundreds_of_thousands",
			amount:   500000,
			expected: "Rp 500.000",
		},
		{
			name:     "millions",
			amount:   1000000,
			expected: "Rp 1.000.000",
		},
		{
			name:     "large_amount",
			amount:   999999999,
			expected: "Rp 999.999.999",
		},
		{
			name:     "zero_amount",
			amount:   0,
			expected: "Rp 0",
		},
		{
			name:     "decimal_truncated",
			amount:   10000.99,
			expected: "Rp 10.000",
		},
		{
			name:     "decimal_rounded_down",
			amount:   10000.49,
			expected: "Rp 10.000",
		},
		{
			name:     "single_digit",
			amount:   7,
			expected: "Rp 7",
		},
		{
			name:     "negative_amount",
			amount:   -5000,
			expected: "Rp -5.000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRupiahWithoutDecimal(tt.amount)
			assert.Equal(t, tt.expected, result)
		})
	}
}
