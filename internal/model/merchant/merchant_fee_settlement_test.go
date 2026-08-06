package merchant

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestSettlementConfig_ValidateRequest(t *testing.T) {
	tests := []struct {
		name        string
		config      SettlementConfig
		expectError bool
	}{
		{
			name: "INSTANT type - valid",
			config: SettlementConfig{
				Type: constant.SettlementTypeInstant,
			},
			expectError: false,
		},
		{
			name: "T+1 type - valid",
			config: SettlementConfig{
				Type: "T+1",
			},
			expectError: false,
		},
		{
			name: "T+10 type - valid",
			config: SettlementConfig{
				Type: "T+10",
			},
			expectError: false,
		},
		{
			name: "D+1 type with valid settlement time - valid",
			config: SettlementConfig{
				Type:           "D+1",
				SettlementTime: "15:00:00",
			},
			expectError: false,
		},
		{
			name: "D+1 type with invalid settlement time - error",
			config: SettlementConfig{
				Type:           "D+1",
				SettlementTime: "25:00:00",
			},
			expectError: true,
		},
		{
			name: "D+1 type with malformed settlement time - error",
			config: SettlementConfig{
				Type:           "D+1",
				SettlementTime: "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid type format - error",
			config: SettlementConfig{
				Type: "INVALID",
			},
			expectError: true,
		},
		{
			name: "T+0 type - error (must be >= 1)",
			config: SettlementConfig{
				Type: "T+0",
			},
			expectError: true,
		},
		{
			name: "D+1 type with CutOff - should set ExecutionTime",
			config: SettlementConfig{
				Type:           "D+1",
				SettlementTime: "10:00:00",
				CutOff: &SettlementConfigCutOff{
					Deferral: SettlementConfigCutOffDeferral{
						ExecutionTime: "00:00:00",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateRequest()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSettlementConfig_GetSettlementDay(t *testing.T) {
	tests := []struct {
		name     string
		config   SettlementConfig
		expected int
	}{
		{
			name:     "INSTANT type - returns 0",
			config:   SettlementConfig{Type: constant.SettlementTypeInstant},
			expected: 0,
		},
		{
			name:     "T+1 - returns 1",
			config:   SettlementConfig{Type: "T+1"},
			expected: 1,
		},
		{
			name:     "T+7 - returns 7",
			config:   SettlementConfig{Type: "T+7"},
			expected: 7,
		},
		{
			name:     "T+30 - returns 30",
			config:   SettlementConfig{Type: "T+30"},
			expected: 30,
		},
		{
			name:     "D+1 - returns 1",
			config:   SettlementConfig{Type: "D+1"},
			expected: 1,
		},
		{
			name:     "D+14 - returns 14",
			config:   SettlementConfig{Type: "D+14"},
			expected: 14,
		},
		{
			name:     "invalid type - returns 0",
			config:   SettlementConfig{Type: "INVALID"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetSettlementDay()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSettlementConfig_GetSettlementTime(t *testing.T) {
	processTime := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)

	tests := []struct {
		name         string
		config       SettlementConfig
		processTime  time.Time
		expectError  bool
		checkResult  func(t *testing.T, result time.Time)
	}{
		{
			name:        "INSTANT type - returns zero time",
			config:      SettlementConfig{Type: constant.SettlementTypeInstant},
			processTime: processTime,
			checkResult: func(t *testing.T, result time.Time) {
				assert.True(t, result.IsZero())
			},
		},
		{
			name: "T+1 without cutoff - returns process time + 1 day",
			config: SettlementConfig{
				Type: "T+1",
			},
			processTime: processTime,
			checkResult: func(t *testing.T, result time.Time) {
				assert.Equal(t, 16, result.Day())
				assert.Equal(t, time.January, result.Month())
				assert.Equal(t, 2024, result.Year())
			},
		},
		{
			name: "T+1 with cutoff not in window - returns execution time",
			config: SettlementConfig{
				Type: "T+1",
				CutOff: &SettlementConfigCutOff{
					Window: SettlementConfigCutOffWindow{
						StartTime: "08:00:00",
						EndTime:   "12:00:00",
					},
					Deferral: SettlementConfigCutOffDeferral{
						OffsetDays:    1,
						ExecutionTime: "09:00:00",
					},
				},
			},
			processTime: time.Date(2024, 1, 15, 14, 0, 0, 0, loc), // 14:00 - outside window
			checkResult: func(t *testing.T, result time.Time) {
				assert.Equal(t, 16, result.Day())
				assert.Equal(t, 9, result.Hour())
				assert.Equal(t, 0, result.Minute())
			},
		},
		{
			name: "T+1 with cutoff in window - adds offset days",
			config: SettlementConfig{
				Type: "T+1",
				CutOff: &SettlementConfigCutOff{
					Window: SettlementConfigCutOffWindow{
						StartTime: "08:00:00",
						EndTime:   "18:00:00",
					},
					Deferral: SettlementConfigCutOffDeferral{
						OffsetDays:    1,
						ExecutionTime: "09:00:00",
					},
				},
			},
			processTime: time.Date(2024, 1, 15, 10, 0, 0, 0, loc), // 10:00 - inside window
			checkResult: func(t *testing.T, result time.Time) {
				assert.Equal(t, 17, result.Day()) // 15 + 1 (T+1) + 1 (offset)
				assert.Equal(t, 9, result.Hour())
			},
		},
		{
			name: "D+1 uses SettlementTime instead of CutOff.ExecutionTime (outside cutoff window)",
			config: SettlementConfig{
				Type:           "D+1",
				SettlementTime: "14:30:00",
				CutOff: &SettlementConfigCutOff{
					Window: SettlementConfigCutOffWindow{
						StartTime: "08:00:00",
						EndTime:   "12:00:00",
					},
					Deferral: SettlementConfigCutOffDeferral{
						OffsetDays:    1,
						ExecutionTime: "09:00:00",
					},
				},
			},
			processTime: time.Date(2024, 1, 15, 14, 0, 0, 0, loc), // 14:00 - outside window, no offset added
			checkResult: func(t *testing.T, result time.Time) {
				assert.Equal(t, 16, result.Day())
				assert.Equal(t, 14, result.Hour())
				assert.Equal(t, 30, result.Minute())
			},
		},
		{
			name: "D+1 with cutoff in window - adds offset days",
			config: SettlementConfig{
				Type:           "D+1",
				SettlementTime: "14:30:00",
				CutOff: &SettlementConfigCutOff{
					Window: SettlementConfigCutOffWindow{
						StartTime: "08:00:00",
						EndTime:   "18:00:00",
					},
					Deferral: SettlementConfigCutOffDeferral{
						OffsetDays:    1,
						ExecutionTime: "09:00:00",
					},
				},
			},
			processTime: time.Date(2024, 1, 15, 10, 0, 0, 0, loc), // 10:00 - inside window, offset added
			checkResult: func(t *testing.T, result time.Time) {
				assert.Equal(t, 17, result.Day()) // 15 + 1 (D+1) + 1 (offset)
				assert.Equal(t, 14, result.Hour())
				assert.Equal(t, 30, result.Minute())
			},
		},
		{
			name: "zero process time - uses current time",
			config: SettlementConfig{
				Type: "T+1",
			},
			processTime: time.Time{}, // zero time
			checkResult: func(t *testing.T, result time.Time) {
				now := time.Now().In(loc)
				assert.True(t, result.After(now) || result.Equal(now))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.config.GetSettlementTime(tt.processTime)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}