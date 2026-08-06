package industry

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestIsValidDigitalStatus(t *testing.T) {
	testCases := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Valid status - Digital",
			status:   "Digital",
			expected: true,
		},
		{
			name:     "Valid status - Non-digital",
			status:   "Non-digital",
			expected: true,
		},
		{
			name:     "Invalid status",
			status:   "Semi-digital",
			expected: false,
		},
		{
			name:     "Case sensitive check",
			status:   "digital",
			expected: false,
		},
		{
			name:     "Empty status",
			status:   "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidDigitalStatus(tc.status)
			assert.Equal(t, tc.expected, result,
				"IsValidDigitalStatus should correctly validate digital status %s", tc.status)

			// Double check with DigitalStatusOptions
			found := false
			for _, validStatus := range constant.DigitalStatusOptions {
				if validStatus == tc.status {
					found = true
					break
				}
			}
			assert.Equal(t, found, result,
				"IsValidDigitalStatus should match manual validation using constant.DigitalStatusOptions")
		})
	}
}

func TestIsValidCountryEntity(t *testing.T) {
	testCases := []struct {
		name     string
		country  string
		expected bool
	}{
		{
			name:     "Valid country - Indonesia",
			country:  "ID",
			expected: true,
		},
		{
			name:     "Valid country - Singapore",
			country:  "SG",
			expected: true,
		},
		{
			name:     "Valid country - United States",
			country:  "US",
			expected: true,
		},
		{
			name:     "Invalid country",
			country:  "XX",
			expected: false,
		},
		{
			name:     "Case sensitive check",
			country:  "id",
			expected: false,
		},
		{
			name:     "Empty country",
			country:  "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidCountryEntity(tc.country)
			assert.Equal(t, tc.expected, result,
				"IsValidCountryEntity should correctly validate country code %s", tc.country)

			// Double check with CountryEntityCodes
			_, found := constant.CountryEntityCodes[tc.country]
			assert.Equal(t, found, result,
				"IsValidCountryEntity should match manual validation using constant.CountryEntityCodes")
		})
	}
}
