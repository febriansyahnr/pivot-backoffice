package util

import "testing"

func TestFormatIDRAmount(t *testing.T) {
	tests := []struct {
		amount   float64
		expected string
	}{
		{1234567.89, "1.234.568"},
		{0, "0"},
		{1000000, "1.000.000"},
	}

	for _, test := range tests {
		result := ConvertFloatToCurrency(test.amount)
		if result != test.expected {
			t.Errorf("ConvertFloatToCurrency(%f) = %s; expected %s", test.amount, result, test.expected)
		}
	}
}

func TestHasDecimal(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected bool
	}{
		{"integer value", 1000, false},
		{"decimal value", 1000.75, true},
		{"zero", 0, false},
		{"negative integer", -1000, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := HasDecimal(test.value)
			if result != test.expected {
				t.Errorf("HasDecimal(%f) = %v; expected %v", test.value, result, test.expected)
			}
		})
	}
}
