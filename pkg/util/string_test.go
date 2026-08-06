package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateDeviceID(t *testing.T) {
	testCases := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "SUCCESS: successfully generate device ID",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userAgent := "testing"
			_ = GenerateDeviceID(userAgent)
		})
	}
}

func TestRandStringBytes(t *testing.T) {
	testCases := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "SUCCESS: successfully generate random string",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = RandStringBytes(10)
		})
	}
}

func TestCreateSlug(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		result  string
		wantErr bool
	}{
		{
			name:    "SUCCESS: successfully create slug",
			input:   "testing 123",
			result:  "TESTING-123",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CreateSlug(tc.input)
			if result != tc.result {
				t.Errorf("CreateSlug() = %v, want %v", result, tc.result)
			}
		})
	}
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase to Title",
			input:    "example string",
			expected: "Example String",
		},
		{
			name:     "Uppercase to Title",
			input:    "EXAMPLE STRING",
			expected: "Example String",
		},
		{
			name:     "Mixed Case to Title",
			input:    "ExAmPlE StRiNg",
			expected: "Example String",
		},
		{
			name:     "With Underscores",
			input:    "example_string_with_underscores",
			expected: "Example String With Underscores",
		},
		{
			name:     "Already in Title Case",
			input:    "Example String",
			expected: "Example String",
		},
		{
			name:     "Single Word",
			input:    "word",
			expected: "Word",
		},
		{
			name:     "Real World Case",
			input:    "PAYMENT - VIRTUAL_ACCOUNT",
			expected: "Payment - Virtual Account",
		},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToTitle(tt.input)
			if result != tt.expected {
				t.Errorf("ToTitle(%q) got %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	randomString, err := GenerateRandomString(10)
	if err != nil {
		t.Errorf("GetJakartaTime() returned an error: %v", err)
	}

	assert.Equal(t, 10, len(randomString))
}

func TestInArray(t *testing.T) {
	tests := []struct {
		arr    []string
		target string
		want   bool
	}{
		{},
		{
			arr:    []string{"ABC", "EFD", "KML"},
			target: "ZXC",
			want:   false,
		},
		{
			arr:    []string{"ABC", "EFD", "ZXC", "KML", "YKJ"},
			target: "ZXC",
			want:   true,
		},
		{
			arr:    []string{"ABC"},
			target: "ABC",
			want:   true,
		},
		{
			arr:    []string{"ZXC", "KML", "YKJ"},
			target: "YKJ",
			want:   true,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, InArray(test.arr, test.target))
	}
}

func TestIsNumericValue(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{},
		{
			value: "1AG",
			want:  false,
		},
		{
			value: "A01",
			want:  false,
		},
		{
			value: "1200",
			want:  true,
		},
		{
			value: "14509876453",
			want:  true,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, IsNumericValue(test.value))
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{
			value: "John Wick",
			want:  "8891449e0a46fcca0bd175a835af040713bd6f803dff54ccac5eaa5c240aa10f",
		},
		{
			value: "123456",
			want:  "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92",
		},
		{
			value: "605d77db-e535-4578-8bd7-624d1ebc8090",
			want:  "34351284c621fcb112f5655df6d55c104d9d5ce06262d26c660b9ae5da0d7b1f",
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, HashString(test.value))
	}
}

func TestGetValueAsString(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string]any
		key          string
		defaultValue string
		expected     string
	}{
		{
			name: "SUCCESS: get existing string value",
			data: map[string]any{
				"test_key": "test_value",
			},
			key:          "test_key",
			defaultValue: "",
			expected:     "test_value",
		},
		{
			name: "SUCCESS: key not found returns default value",
			data: map[string]any{
				"other_key": "test_value",
			},
			key:          "test_key",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name: "SUCCESS: value is not string returns default value",
			data: map[string]any{
				"test_key": 123,
			},
			key:          "test_key",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "SUCCESS: empty map returns default value",
			data:         map[string]any{},
			key:          "test_key",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name: "SUCCESS: nil value returns default value",
			data: map[string]any{
				"test_key": nil,
			},
			key:          "test_key",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetValueAsString(tc.data, tc.key, tc.defaultValue)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTrimLength(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "SUCCESS: string shorter than maxLength",
			input:     "test",
			maxLength: 10,
			expected:  "test",
		},
		{
			name:      "SUCCESS: string equal to maxLength",
			input:     "test1234",
			maxLength: 8,
			expected:  "test1234",
		},
		{
			name:      "SUCCESS: string longer than maxLength",
			input:     "test1234567890",
			maxLength: 5,
			expected:  "67890",
		},
		{
			name:      "SUCCESS: empty string input",
			input:     "",
			maxLength: 5,
			expected:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := TrimLength(tc.input, tc.maxLength)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTrimLengthRight(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "SUCCESS: string more than maxLength",
			input:     "test",
			maxLength: 10,
			expected:  "test",
		},
		{
			name:      "SUCCESS: string equal to maxLength",
			input:     "test1234",
			maxLength: 8,
			expected:  "test1234",
		},
		{
			name:      "SUCCESS: string longer than maxLength",
			input:     "test1234567890",
			maxLength: 5,
			expected:  "test1",
		},
		{
			name:      "SUCCESS: empty string input",
			input:     "",
			maxLength: 5,
			expected:  "",
		},
		{
			name:      "SUCCESS: negative length",
			input:     "test",
			maxLength: -1,
			expected:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := TrimLengthRight(tc.input, tc.maxLength)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRemoveNameExtraSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SUCCESS: Same string",
			input:    "test",
			expected: "test",
		},
		{
			name:     "SUCCESS: Space in the start and end",
			input:    " test1234 ",
			expected: "test1234",
		},
		{
			name:     "SUCCESS: Space in the middle",
			input:    "test 1234",
			expected: "test 1234",
		},

		{
			name:     "SUCCESS: Extra space in the middle",
			input:    "test                      12          34",
			expected: "test 12 34",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := RemoveNameExtraSpace(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
