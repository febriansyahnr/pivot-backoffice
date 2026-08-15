package amlcommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractNodeAttributes(t *testing.T) {
	testCases := []struct {
		name       string
		attributes any
		expected   NodeAttributes
	}{
		{
			name: "SUCCESS: extract complete attributes",
			attributes: map[string]any{
				"dob":               "1990-01-01",
				"name":              "John Doe",
				"score":             95.0,
				"gender":            "MALE",
				"entityType":        "PERSON",
				"hitCategory":       []any{"PEP", "LE", "SAN"},
				"referenceId":       "ref-123-456",
				"placeOfBirth":      "Jakarta",
				"countryLocation":   "Indonesia",
				"registeredCountry": "Indonesia",
			},
			expected: NodeAttributes{
				DOB:               "1990-01-01",
				Name:              "John Doe",
				Score:             95,
				Gender:            "MALE",
				EntityType:        "PERSON",
				HitCategory:       []string{"PEP", "LE", "SAN"},
				ReferenceID:       "ref-123-456",
				PlaceOfBirth:      "Jakarta",
				CountryLocation:   "Indonesia",
				RegisteredCountry: "Indonesia",
			},
		},
		{
			name: "SUCCESS: extract partial attributes",
			attributes: map[string]any{
				"name":   "Jane Smith",
				"gender": "FEMALE",
				"score":  80.5,
			},
			expected: NodeAttributes{
				Name:   "Jane Smith",
				Gender: "FEMALE",
				Score:  80,
			},
		},
		{
			name: "SUCCESS: handle empty hitCategory array",
			attributes: map[string]any{
				"name":        "Test User",
				"hitCategory": []any{},
			},
			expected: NodeAttributes{
				Name:        "Test User",
				HitCategory: []string{},
			},
		},
		{
			name: "SUCCESS: handle mixed hitCategory types",
			attributes: map[string]any{
				"name":        "Mixed User",
				"hitCategory": []any{"PEP", 123, "LE", nil, "SAN"},
			},
			expected: NodeAttributes{
				Name:        "Mixed User",
				HitCategory: []string{"PEP", "", "LE", "", "SAN"},
			},
		},
		{
			name:       "FAIL: nil attributes",
			attributes: nil,
			expected:   NodeAttributes{},
		},
		{
			name:       "FAIL: non-map attributes",
			attributes: "invalid string",
			expected:   NodeAttributes{},
		},
		{
			name: "FAIL: empty map attributes",
			attributes: map[string]any{},
			expected:   NodeAttributes{},
		},
		{
			name: "SUCCESS: handle wrong types gracefully",
			attributes: map[string]any{
				"dob":               123,        // Wrong type (should be string)
				"name":              []string{}, // Wrong type (should be string)
				"score":             "not-a-number", // Wrong type (should be float64)
				"gender":            true,       // Wrong type (should be string)
				"hitCategory":       "not-array", // Wrong type (should be array)
			},
			expected: NodeAttributes{
				// All fields should remain empty due to type mismatches
			},
		},
		{
			name: "SUCCESS: handle score as integer",
			attributes: map[string]any{
				"name":  "Integer Score User",
				"score": 95, // Integer instead of float64
			},
			expected: NodeAttributes{
				Name: "Integer Score User",
				// Score should remain 0 since we expect float64 but got int
			},
		},
		{
			name: "SUCCESS: extract with special characters",
			attributes: map[string]any{
				"name":              "José María Ñoño",
				"placeOfBirth":      "São Paulo",
				"countryLocation":   "República Argentina",
				"registeredCountry": "México D.F.",
				"referenceId":       "ref-åäö-123",
			},
			expected: NodeAttributes{
				Name:              "José María Ñoño",
				PlaceOfBirth:      "São Paulo",
				CountryLocation:   "República Argentina",
				RegisteredCountry: "México D.F.",
				ReferenceID:       "ref-åäö-123",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function
			result := ExtractNodeAttributes(tc.attributes)

			// Assertions
			assert.Equal(t, tc.expected.DOB, result.DOB)
			assert.Equal(t, tc.expected.Name, result.Name)
			assert.Equal(t, tc.expected.Score, result.Score)
			assert.Equal(t, tc.expected.Gender, result.Gender)
			assert.Equal(t, tc.expected.EntityType, result.EntityType)
			assert.Equal(t, tc.expected.ReferenceID, result.ReferenceID)
			assert.Equal(t, tc.expected.PlaceOfBirth, result.PlaceOfBirth)
			assert.Equal(t, tc.expected.CountryLocation, result.CountryLocation)
			assert.Equal(t, tc.expected.RegisteredCountry, result.RegisteredCountry)
			assert.Equal(t, tc.expected.HitCategory, result.HitCategory)
		})
	}
}