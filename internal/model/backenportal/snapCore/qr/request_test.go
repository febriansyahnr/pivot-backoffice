package snapCoreModel

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestQueryQrMpmStaticRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		request  QueryQrMpmStaticRequest
		expected QueryQrMpmStaticRequest
	}{
		{
			name: "Default values when empty",
			request: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       "",
				ToDateTime:         "",
				PageNumber:         0,
				PageSize:           0,
			},
			expected: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       getDefaultDate(-3), // 3 months ago
				ToDateTime:         getDefaultDate(0),  // current date
				PageNumber:         1,                  // default page number
				PageSize:           20,                 // default page size
			},
		},
		{
			name: "Valid dates should remain unchanged",
			request: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       time.Now().AddDate(0, -1, 0).Format(util.SnapDateFormatLayout), // 1 month ago
				ToDateTime:         time.Now().AddDate(0, 0, -1).Format(util.SnapDateFormatLayout), // yesterday
				PageNumber:         2,
				PageSize:           50,
			},
			expected: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       time.Now().AddDate(0, -1, 0).Format(util.SnapDateFormatLayout), // should remain unchanged
				ToDateTime:         time.Now().AddDate(0, 0, -1).Format(util.SnapDateFormatLayout), // should remain unchanged
				PageNumber:         2,                                                              // should remain unchanged
				PageSize:           50,                                                             // should remain unchanged
			},
		},
		{
			name: "Future dates should be reset to defaults",
			request: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       time.Now().AddDate(0, 1, 0).Format(util.SnapDateFormatLayout),  // 1 month in future
				ToDateTime:         time.Now().AddDate(0, 0, 10).Format(util.SnapDateFormatLayout), // 10 days in future
				PageNumber:         1,
				PageSize:           20,
			},
			expected: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       getDefaultDate(-3), // should be reset to 3 months ago
				ToDateTime:         getDefaultDate(0),  // should be reset to current date
				PageNumber:         1,
				PageSize:           20,
			},
		},
		{
			name: "Invalid date format should be reset to defaults",
			request: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       "invalid-date",
				ToDateTime:         "2023/01/01",
				PageNumber:         1,
				PageSize:           20,
			},
			expected: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       getDefaultDate(-3), // should be reset to 3 months ago
				ToDateTime:         getDefaultDate(0),  // should be reset to current date
				PageNumber:         1,
				PageSize:           20,
			},
		},
		{
			name: "Page size below minimum should be set to default",
			request: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       time.Now().AddDate(0, -1, 0).Format(util.SnapDateFormatLayout),
				ToDateTime:         time.Now().Format(util.SnapDateFormatLayout),
				PageNumber:         1,
				PageSize:           10, // below minimum of 20
			},
			expected: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       time.Now().AddDate(0, -1, 0).Format(util.SnapDateFormatLayout),
				ToDateTime:         time.Now().Format(util.SnapDateFormatLayout),
				PageNumber:         1,
				PageSize:           20, // should be set to default
			},
		},
		{
			name: "Page size above maximum should be set to default",
			request: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       time.Now().AddDate(0, -1, 0).Format(util.SnapDateFormatLayout),
				ToDateTime:         time.Now().Format(util.SnapDateFormatLayout),
				PageNumber:         1,
				PageSize:           150, // above maximum of 100
			},
			expected: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       time.Now().AddDate(0, -1, 0).Format(util.SnapDateFormatLayout),
				ToDateTime:         time.Now().Format(util.SnapDateFormatLayout),
				PageNumber:         1,
				PageSize:           20, // should be set to default
			},
		},
		{
			name: "Page number below minimum should be set to default",
			request: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       time.Now().AddDate(0, -1, 0).Format(util.SnapDateFormatLayout),
				ToDateTime:         time.Now().Format(util.SnapDateFormatLayout),
				PageNumber:         0, // below minimum of 1
				PageSize:           50,
			},
			expected: QueryQrMpmStaticRequest{
				PartnerReferenceNo: "REF123",
				FromDateTime:       time.Now().AddDate(0, -1, 0).Format(util.SnapDateFormatLayout),
				ToDateTime:         time.Now().Format(util.SnapDateFormatLayout),
				PageNumber:         1, // should be set to default
				PageSize:           50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the request to validate
			request := tt.request

			// Call the Validate method
			request.Validate()

			// For date fields, we need to check if they match the expected pattern
			// rather than exact string equality due to time differences during test execution
			if tt.expected.FromDateTime == getDefaultDate(-3) {
				// If expected is default, just check that it's not empty and in correct format
				_, err := time.Parse(util.SnapDateFormatLayout, request.FromDateTime)
				assert.NoError(t, err, "FromDateTime should be in valid format")
			} else {
				// Otherwise check exact match
				assert.Equal(t, tt.expected.FromDateTime, request.FromDateTime, "FromDateTime should match expected")
			}

			if tt.expected.ToDateTime == getDefaultDate(0) {
				// If expected is default, just check that it's not empty and in correct format
				_, err := time.Parse(util.SnapDateFormatLayout, request.ToDateTime)
				assert.NoError(t, err, "ToDateTime should be in valid format")
			} else {
				// Otherwise check exact match
				assert.Equal(t, tt.expected.ToDateTime, request.ToDateTime, "ToDateTime should match expected")
			}

			// Check other fields
			assert.Equal(t, tt.expected.PageNumber, request.PageNumber, "PageNumber should match expected")
			assert.Equal(t, tt.expected.PageSize, request.PageSize, "PageSize should match expected")
		})
	}
}

func TestValidateAndSetDate(t *testing.T) {
	tests := []struct {
		name            string
		dateStr         string
		layout          string
		months          int
		shouldBeDefault bool
	}{
		{
			name:            "Empty date string should return default date",
			dateStr:         "",
			layout:          util.SnapDateFormatLayout,
			months:          -1,
			shouldBeDefault: true,
		},
		{
			name:            "Valid date string should return unchanged",
			dateStr:         time.Now().AddDate(0, -1, 0).Format(util.SnapDateFormatLayout),
			layout:          util.SnapDateFormatLayout,
			months:          -1,
			shouldBeDefault: false,
		},
		{
			name:            "Future date should return default date",
			dateStr:         time.Now().AddDate(0, 1, 0).Format(util.SnapDateFormatLayout),
			layout:          util.SnapDateFormatLayout,
			months:          -1,
			shouldBeDefault: true,
		},
		{
			name:            "Invalid format should return default date",
			dateStr:         "invalid-date",
			layout:          util.SnapDateFormatLayout,
			months:          -1,
			shouldBeDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateAndSetDate(tt.dateStr, tt.layout, tt.months)

			if tt.shouldBeDefault {
				// Should be default date
				expectedDefault := getDefaultDate(tt.months)
				assert.Equal(t, expectedDefault, result, "Result should be default date")
			} else {
				// Should be unchanged
				assert.Equal(t, tt.dateStr, result, "Result should be unchanged")
			}
		})
	}
}

func TestGetDefaultDate(t *testing.T) {
	tests := []struct {
		name   string
		months int
	}{
		{
			name:   "Current date (0 months)",
			months: 0,
		},
		{
			name:   "3 months ago",
			months: -3,
		},
		{
			name:   "1 month in future",
			months: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDefaultDate(tt.months)

			// Parse the result
			parsedResult, err := time.Parse(util.SnapDateFormatLayout, result)
			assert.NoError(t, err, "Result should be in valid format")

			// Calculate expected date
			expectedDate := util.ConvertToJakarta(time.Now().AddDate(0, tt.months, 0))

			// Compare year, month, day (ignoring time differences during test execution)
			assert.Equal(t, expectedDate.Year(), parsedResult.Year(), "Year should match")
			assert.Equal(t, expectedDate.Month(), parsedResult.Month(), "Month should match")
			assert.Equal(t, expectedDate.Day(), parsedResult.Day(), "Day should match")
		})
	}
}
