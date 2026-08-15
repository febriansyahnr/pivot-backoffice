package countryModel_test

import (
	"database/sql"
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/country"
	"github.com/stretchr/testify/assert"
)

func TestCountry_ToResponse(t *testing.T) {
	// Create base time for consistent testing
	baseTime := time.Date(2023, 12, 25, 10, 30, 45, 123456789, time.UTC)
	updatedTime := time.Date(2023, 12, 26, 15, 45, 30, 987654321, time.UTC)

	tests := []struct {
		name     string
		country  *Country
		expected *CountryResponse
	}{
		{
			name: "complete_country_data",
			country: &Country{
				Code:      "ID",
				Name:      "Indonesia",
				NameID:    "Indonesia",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "ID",
				Name:      "Indonesia",
				NameID:    "Indonesia",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
			},
		},
		{
			name: "country_with_different_name_and_name_id",
			country: &Country{
				Code:      "US",
				Name:      "United States",
				NameID:    "Amerika Serikat",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "US",
				Name:      "United States",
				NameID:    "Amerika Serikat",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
		},
		{
			name: "country_with_special_characters",
			country: &Country{
				Code:      "DE",
				Name:      "Deutschland",
				NameID:    "Jérman",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "DE",
				Name:      "Deutschland",
				NameID:    "Jérman",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
			},
		},
		{
			name: "country_with_unicode_characters",
			country: &Country{
				Code:      "JP",
				Name:      "Japan",
				NameID:    "日本",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "JP",
				Name:      "Japan",
				NameID:    "日本",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
			},
		},
		{
			name: "country_with_empty_strings",
			country: &Country{
				Code:      "",
				Name:      "",
				NameID:    "",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "",
				Name:      "",
				NameID:    "",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
			},
		},
		{
			name: "country_with_long_names",
			country: &Country{
				Code:      "XX",
				Name:      "Very Long Country Name That Exceeds Normal Length Requirements",
				NameID:    "Nama Negara Yang Sangat Panjang Dan Melebihi Persyaratan Panjang Normal",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "XX",
				Name:      "Very Long Country Name That Exceeds Normal Length Requirements",
				NameID:    "Nama Negara Yang Sangat Panjang Dan Melebihi Persyaratan Panjang Normal",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
			},
		},
		{
			name: "country_with_special_iso_code",
			country: &Country{
				Code:      "GB",
				Name:      "United Kingdom",
				NameID:    "Kerajaan Bersatu",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "GB",
				Name:      "United Kingdom",
				NameID:    "Kerajaan Bersatu",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
			},
		},
		{
			name: "country_with_deleted_at_set",
			country: &Country{
				Code:      "FR",
				Name:      "France",
				NameID:    "Perancis",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
				DeletedAt: sql.NullTime{Time: baseTime.Add(24 * time.Hour), Valid: true},
			},
			expected: &CountryResponse{
				Code:      "FR",
				Name:      "France",
				NameID:    "Perancis",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
			},
		},
		{
			name: "country_with_same_created_and_updated_time",
			country: &Country{
				Code:      "CA",
				Name:      "Canada",
				NameID:    "Kanada",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "CA",
				Name:      "Canada",
				NameID:    "Kanada",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
		},
		{
			name: "country_with_timezone_edge_case",
			country: &Country{
				Code:      "AU",
				Name:      "Australia",
				NameID:    "Australia",
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 12, 31, 23, 59, 59, 999999999, time.UTC),
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "AU",
				Name:      "Australia",
				NameID:    "Australia",
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 12, 31, 23, 59, 59, 999999999, time.UTC),
			},
		},
		{
			name: "country_with_numeric_in_name",
			country: &Country{
				Code:      "XX",
				Name:      "Country 123",
				NameID:    "Negara 123",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "XX",
				Name:      "Country 123",
				NameID:    "Negara 123",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
			},
		},
		{
			name: "country_with_spaces_and_punctuation",
			country: &Country{
				Code:      "XX",
				Name:      "St. Vincent & the Grenadines",
				NameID:    "Santo Vincent & Grenadines",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
				DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
			},
			expected: &CountryResponse{
				Code:      "XX",
				Name:      "St. Vincent & the Grenadines",
				NameID:    "Santo Vincent & Grenadines",
				CreatedAt: baseTime,
				UpdatedAt: updatedTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.country.ToResponse()
			assert.Equal(t, tt.expected.Code, result.Code)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.NameID, result.NameID)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)

			// Verify that DeletedAt is not included in response
			assert.NotNil(t, result)
			assert.IsType(t, &CountryResponse{}, result)
		})
	}
}

func TestCountry_ToResponse_NilPointer(t *testing.T) {
	// Test edge case where country pointer might be nil
	var country *Country = nil

	// This should panic as expected in Go when calling method on nil pointer
	assert.Panics(t, func() {
		country.ToResponse()
	}, "Calling ToResponse on nil Country pointer should panic")
}

func TestCountry_ToResponse_FieldMapping(t *testing.T) {
	// Test to ensure all fields are correctly mapped and DeletedAt is excluded
	baseTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	deletedTime := time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)

	country := &Country{
		Code:      "TEST",
		Name:      "Test Country",
		NameID:    "Negara Test",
		CreatedAt: baseTime,
		UpdatedAt: baseTime.Add(1 * time.Hour),
		DeletedAt: sql.NullTime{Time: deletedTime, Valid: true},
	}

	response := country.ToResponse()

	// Verify all fields are mapped correctly
	assert.Equal(t, "TEST", response.Code)
	assert.Equal(t, "Test Country", response.Name)
	assert.Equal(t, "Negara Test", response.NameID)
	assert.Equal(t, baseTime, response.CreatedAt)
	assert.Equal(t, baseTime.Add(1*time.Hour), response.UpdatedAt)

	// Verify response struct doesn't have DeletedAt field
	// This is implicit in the struct definition, but we verify the response is clean
	assert.NotNil(t, response)
	assert.IsType(t, &CountryResponse{}, response)
}
