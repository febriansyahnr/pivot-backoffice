package util

import "testing"

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    interface{}
		item     interface{}
		expected bool
	}{
		{"String Found", []string{"apple", "banana", "cherry"}, "banana", true},
		{"String Not Found", []string{"apple", "banana", "cherry"}, "grape", false},
		{"Int Found", []int{1, 2, 3, 4, 5}, 3, true},
		{"Int Not Found", []int{1, 2, 3, 4, 5}, 10, false},
		{"Float Found", []float64{1.1, 2.2, 3.3, 4.4}, 2.2, true},
		{"Float Not Found", []float64{1.1, 2.2, 3.3, 4.4}, 5.5, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch v := test.slice.(type) {
			case []string:
				result := Contains(v, test.item.(string))
				if result != test.expected {
					t.Errorf("Expected %v, got %v", test.expected, result)
				}
			case []int:
				result := Contains(v, test.item.(int))
				if result != test.expected {
					t.Errorf("Expected %v, got %v", test.expected, result)
				}
			case []float64:
				result := Contains(v, test.item.(float64))
				if result != test.expected {
					t.Errorf("Expected %v, got %v", test.expected, result)
				}
			default:
				t.Errorf("Unsupported type %T", v)
			}
		})
	}
}

func TestContainsPrefix(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		path     string
		expected bool
	}{
		{
			name:     "Path matches exact prefix",
			slice:    []string{"/health-check", "/ping"},
			path:     "/api/v1/health-check",
			expected: true,
		},
		{
			name:     "Path does not match any prefix",
			slice:    []string{"/health-check", "/ping"},
			path:     "/api/v1/status",
			expected: false,
		},
		{
			name:     "Empty slice",
			slice:    []string{},
			path:     "/health-check/api",
			expected: false,
		},
		{
			name:     "Path matches last prefix",
			slice:    []string{"/api/v1/", "/health-check"},
			path:     "/api/v1/resource",
			expected: true,
		},
		{
			name:     "Path matches with longer prefix",
			slice:    []string{"/api/", "/health"},
			path:     "/api/v1/resource",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsPrefix(tt.slice, tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
