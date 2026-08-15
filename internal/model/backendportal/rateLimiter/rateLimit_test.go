package ratelimiter

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Save the original generator and restore it after the test
	originalGenerator := defaultUUIDGenerator
	defer func() {
		defaultUUIDGenerator = originalGenerator
	}()

	// Test cases
	testCases := []struct {
		name    string
		request *CreateRateLimitConfiguration
	}{
		{
			name: "Valid configuration",
			request: &CreateRateLimitConfiguration{
				MerchantID:        "merchant-123",
				Limit:             100,
				Order:             1,
				Time:              constant.RateLimitConfigTimeSecond,
				Variable:          "ip",
				VariableValue:     "127.0.0.1",
				VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
				Description:       "Test rate limit",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the generator to the default for each test case
			defaultUUIDGenerator = originalGenerator

			// Call the function
			result, err := New(tc.request)

			// Assertions
			assert.NoError(t, err, "New should not return an error")
			assert.NotNil(t, result, "Result should not be nil")

			// Verify UUID is valid
			_, uuidErr := uuid.Parse(result.UUID)
			assert.NoError(t, uuidErr, "UUID should be valid")

			// Verify other fields
			assert.Equal(t, tc.request.MerchantID, result.MerchantID)
			assert.Equal(t, tc.request.Limit, result.Limit)
			assert.Equal(t, tc.request.Order, result.Order)
			assert.Equal(t, tc.request.Time, result.Time)
			assert.Equal(t, tc.request.Variable, result.Variable)
			assert.Equal(t, tc.request.VariableValue, result.VariableValue)
			assert.Equal(t, tc.request.VariableMatchType, result.VariableMatchType)
			assert.Equal(t, tc.request.Description, result.Description)
			assert.Equal(t, constant.StatusActive, result.Status)

			// Verify timestamps
			assert.WithinDuration(t, time.Now().UTC(), result.CreatedAt, 2*time.Second)
			assert.WithinDuration(t, time.Now().UTC(), result.UpdatedAt, 2*time.Second)
		})
	}
}

// TestNewErrorHandling tests the error handling of New
func TestNewErrorHandling(t *testing.T) {
	// Save the original generator and restore it after the test
	originalGenerator := defaultUUIDGenerator
	defer func() {
		defaultUUIDGenerator = originalGenerator
	}()

	// Replace the UUID generator with one that returns an error
	defaultUUIDGenerator = func() (uuid.UUID, error) {
		return uuid.Nil, errors.New("simulated UUID generation error")
	}

	// Create a valid request
	request := &CreateRateLimitConfiguration{
		MerchantID:        "merchant-123",
		Limit:             100,
		Order:             1,
		Time:              constant.RateLimitConfigTimeSecond,
		Variable:          "ip",
		VariableValue:     "127.0.0.1",
		VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
		Description:       "Test rate limit",
	}

	// Call the function with the mock generator
	result, err := New(request)

	// Verify that the error is propagated and the result is nil
	assert.Error(t, err, "Function should return an error when UUID generation fails")
	assert.Nil(t, result, "Result should be nil when an error occurs")
	assert.Contains(t, err.Error(), "simulated UUID generation error", "Error message should be propagated")
}

func TestRateLimitConfig_IsExactType(t *testing.T) {
	r := &MerchantRateLimitConfig{
		VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
	}

	assert.True(t, r.IsExactType())
}

func TestRateLimitConfig_IsContainsType(t *testing.T) {
	r := &MerchantRateLimitConfig{
		VariableMatchType: constant.RateLimitConfigVariableMatchTypeContains,
	}

	assert.True(t, r.IsContainsType())
}

func TestRateLimitConfig_IsPrefixType(t *testing.T) {
	r := &MerchantRateLimitConfig{
		VariableMatchType: constant.RateLimitConfigVariableMatchTypePrefix,
	}

	assert.True(t, r.IsPrefixType())
}

func TestRateLimitConfig_GetDuration(t *testing.T) {
	tests := []struct {
		name     string
		time     string
		expected time.Duration
	}{
		{
			name:     "Second duration",
			time:     constant.RateLimitConfigTimeSecond,
			expected: constant.RateLimitConfigTimeSecondDuration,
		},
		{
			name:     "Minute duration",
			time:     constant.RateLimitConfigTimeMinute,
			expected: constant.RateLimitConfigTimeMinuteDuration,
		},
		{
			name:     "Hour duration",
			time:     constant.RateLimitConfigTimeHour,
			expected: constant.RateLimitConfigTimeHourDuration,
		},
		{
			name:     "Daily duration",
			time:     constant.RateLimitConfigTimeDaily,
			expected: constant.RateLimitConfigTimeDailyDuration,
		},
		{
			name:     "Default duration",
			time:     "UNKNOWN",
			expected: constant.RateLimitConfigTimeSecondDuration,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := &MerchantRateLimitConfig{
				Time: test.time,
			}
			assert.Equal(t, test.expected, r.GetDuration())
		})
	}
}

// TestRateLimitConfiguration_Update tests the Update method of RateLimitConfiguration
func TestRateLimitConfiguration_Update(t *testing.T) {
	// Create an initial configuration
	initialConfig := &RateLimitConfiguration{
		UUID:              "test-uuid",
		MerchantID:        "old-merchant-id",
		Limit:             50,
		Order:             0,
		Time:              constant.RateLimitConfigTimeSecond,
		Variable:          "IP",
		VariableValue:     "old-value",
		VariableMatchType: constant.RateLimitConfigVariableMatchTypeExact,
		Status:            constant.StatusActive,
		Description:       "Old description",
		CreatedAt:         time.Now().UTC().Add(-24 * time.Hour), // 1 day ago
		UpdatedAt:         time.Now().UTC().Add(-24 * time.Hour), // 1 day ago
	}

	// Create an update request
	updateRequest := &UpdateRateLimitConfiguration{
		ID:                "test-uuid",
		MerchantID:        "new-merchant-id",
		Limit:             100,
		Order:             1,
		Time:              constant.RateLimitConfigTimeMinute,
		Variable:          "PATH",
		VariableValue:     "new-value",
		VariableMatchType: constant.RateLimitConfigVariableMatchTypeContains,
		Status:            constant.StatusInactive,
		Description:       "New description",
	}

	// Record the time before update
	beforeUpdate := time.Now().UTC()

	// Call the Update method
	err := initialConfig.Update(updateRequest)

	// Verify no error occurred
	assert.NoError(t, err, "Update should not return an error")

	// Verify fields were updated correctly
	assert.Equal(t, updateRequest.MerchantID, initialConfig.MerchantID, "MerchantID should be updated")
	assert.Equal(t, updateRequest.Limit, initialConfig.Limit, "Limit should be updated")
	assert.Equal(t, updateRequest.Order, initialConfig.Order, "Order should be updated")
	assert.Equal(t, updateRequest.Time, initialConfig.Time, "Time should be updated")
	assert.Equal(t, updateRequest.Variable, initialConfig.Variable, "Variable should be updated")
	assert.Equal(t, updateRequest.VariableValue, initialConfig.VariableValue, "VariableValue should be updated")
	assert.Equal(t, updateRequest.VariableMatchType, initialConfig.VariableMatchType, "VariableMatchType should be updated")
	assert.Equal(t, updateRequest.Status, initialConfig.Status, "Status should be updated")
	assert.Equal(t, updateRequest.Description, initialConfig.Description, "Description should be updated")

	// Verify UUID was not changed
	assert.Equal(t, "test-uuid", initialConfig.UUID, "UUID should not be changed")

	// Verify CreatedAt was not changed
	assert.True(t, initialConfig.CreatedAt.Before(beforeUpdate), "CreatedAt should not be changed")

	// Verify UpdatedAt was updated
	assert.True(t, initialConfig.UpdatedAt.After(beforeUpdate) || initialConfig.UpdatedAt.Equal(beforeUpdate),
		"UpdatedAt should be updated to current time or later")
}
