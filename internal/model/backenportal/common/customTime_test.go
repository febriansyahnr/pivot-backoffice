package commonModel

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCustomNullTimeMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    CustomNullTime
		expected string
		wantErr  bool
	}{
		{
			name:     "invalid_null_time",
			input:    CustomNullTime{sql.NullTime{Valid: false}},
			expected: `{"Valid":false,"Time":""}`,
			wantErr:  false,
		},
		{
			name:     "valid_time_utc",
			input:    CustomNullTime{sql.NullTime{Time: time.Date(2023, 5, 27, 0, 0, 0, 0, time.UTC), Valid: true}},
			expected: `{"Valid":true,"Time":"2023-05-27T00:00:00Z"}`,
			wantErr:  false,
		},
		{
			name:     "valid_time_with_nanoseconds",
			input:    CustomNullTime{sql.NullTime{Time: time.Date(2023, 5, 27, 12, 30, 45, 123456789, time.UTC), Valid: true}},
			expected: `{"Valid":true,"Time":"2023-05-27T12:30:45Z"}`,
			wantErr:  false,
		},
		{
			name:     "valid_time_year_boundary",
			input:    CustomNullTime{sql.NullTime{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}},
			expected: `{"Valid":true,"Time":"2000-01-01T00:00:00Z"}`,
			wantErr:  false,
		},
		{
			name:     "valid_time_leap_year",
			input:    CustomNullTime{sql.NullTime{Time: time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC), Valid: true}},
			expected: `{"Valid":true,"Time":"2024-02-29T23:59:59Z"}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.MarshalJSON()
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(result))
		})
	}
}

func TestCustomNullTimeScan(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		expectedValid bool
		expectedTime  time.Time
		wantErr       bool
	}{
		{
			name:          "nil_input",
			input:         nil,
			expectedValid: false,
			expectedTime:  time.Time{},
			wantErr:       false,
		},
		{
			name:          "valid_time_utc",
			input:         time.Date(2023, 5, 27, 0, 0, 0, 0, time.UTC),
			expectedValid: true,
			expectedTime:  time.Date(2023, 5, 27, 0, 0, 0, 0, time.UTC),
			wantErr:       false,
		},
		{
			name:          "valid_time_with_nanoseconds",
			input:         time.Date(2023, 5, 27, 12, 30, 45, 123456789, time.UTC),
			expectedValid: true,
			expectedTime:  time.Date(2023, 5, 27, 12, 30, 45, 123456789, time.UTC),
			wantErr:       false,
		},
		{
			name:          "valid_time_different_timezone",
			input:         time.Date(2023, 5, 27, 12, 0, 0, 0, time.FixedZone("JST", 9*3600)),
			expectedValid: true,
			expectedTime:  time.Date(2023, 5, 27, 12, 0, 0, 0, time.FixedZone("JST", 9*3600)),
			wantErr:       false,
		},
		{
			name:          "zero_time",
			input:         time.Time{},
			expectedValid: true,
			expectedTime:  time.Time{},
			wantErr:       false,
		},
		{
			name:          "invalid_string_input",
			input:         "invalid input",
			expectedValid: false,
			expectedTime:  time.Time{},
			wantErr:       true,
		},
		{
			name:          "invalid_int_input",
			input:         12345,
			expectedValid: false,
			expectedTime:  time.Time{},
			wantErr:       true,
		},
		{
			name:          "invalid_bool_input",
			input:         true,
			expectedValid: false,
			expectedTime:  time.Time{},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := &CustomNullTime{}
			err := ct.Scan(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedValid, ct.Valid)
			if tt.expectedValid {
				assert.Equal(t, tt.expectedTime, ct.Time)
			}
		})
	}
}
