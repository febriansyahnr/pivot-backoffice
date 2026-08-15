package reconciliation

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestNewReconciliation(t *testing.T) {
	// Save the original generator and restore it after the test
	originalGenerator := defaultUUIDGenerator
	defer func() {
		defaultUUIDGenerator = originalGenerator
	}()

	// Test cases
	testCases := []struct {
		name      string
		createdBy string
		filePath  string
	}{
		{
			name:      "Valid input with user ID",
			createdBy: "user-123",
			filePath:  "/path/to/file.csv",
		},
		{
			name:      "Valid input with empty user ID",
			createdBy: "",
			filePath:  "/path/to/file.csv",
		},
		{
			name:      "Valid input with empty file path",
			createdBy: "user-123",
			filePath:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the generator to the default for each test case
			defaultUUIDGenerator = originalGenerator

			// Call the function
			result, err := NewReconciliation("PAYMENT", tc.createdBy, tc.filePath)

			// Assertions
			assert.NoError(t, err, "NewReconciliation should not return an error")
			assert.NotNil(t, result, "Result should not be nil")

			// Verify UUID is valid
			_, uuidErr := uuid.Parse(result.UUID)
			assert.NoError(t, uuidErr, "UUID should be valid")

			// Verify other fields
			assert.Equal(t, tc.createdBy, result.CreatedBy, "CreatedBy should match input")
			assert.Equal(t, tc.filePath, result.FilePath, "FilePath should match input")
			assert.Equal(t, constant.StatusPending, result.Status, "Status should be PENDING")
			assert.Equal(t, "", result.ResultFilePath, "ResultFilePath should be empty")

			// Verify timestamps
			assert.WithinDuration(t, time.Now(), result.CreatedAt, 2*time.Second, "CreatedAt should be close to current time")
			assert.WithinDuration(t, time.Now(), result.UpdatedAt, 2*time.Second, "UpdatedAt should be close to current time")
			assert.Equal(t, result.CreatedAt, result.UpdatedAt, "CreatedAt and UpdatedAt should be equal")
		})
	}
}

// TestNewReconciliationUUIDUniqueness tests that each call to NewReconciliation generates a unique UUID
func TestNewReconciliationUUIDUniqueness(t *testing.T) {
	// Save the original generator and restore it after the test
	originalGenerator := defaultUUIDGenerator
	defer func() {
		defaultUUIDGenerator = originalGenerator
	}()

	// Generate multiple reconciliation objects
	count := 10
	uuids := make([]string, count)

	for i := 0; i < count; i++ {
		reconciliation, err := NewReconciliation("PAYMENT", "user-123", "/path/to/file.csv")
		assert.NoError(t, err, "NewReconciliation should not return an error")
		uuids[i] = reconciliation.UUID
	}

	// Check for uniqueness
	uniqueUUIDs := make(map[string]bool)
	for _, id := range uuids {
		// If the UUID is already in the map, it's a duplicate
		assert.False(t, uniqueUUIDs[id], "UUID should be unique: %s", id)
		uniqueUUIDs[id] = true
	}
	assert.Equal(t, count, len(uniqueUUIDs), "All UUIDs should be unique")
}

// TestNewReconciliationErrorHandling tests the error handling of NewReconciliation
func TestNewReconciliationErrorHandling(t *testing.T) {
	// Save the original generator and restore it after the test
	originalGenerator := defaultUUIDGenerator
	defer func() {
		defaultUUIDGenerator = originalGenerator
	}()

	// Replace the UUID generator with one that returns an error
	defaultUUIDGenerator = func() (uuid.UUID, error) {
		return uuid.Nil, errors.New("simulated UUID generation error")
	}

	// Call the function with the mock generator
	result, err := NewReconciliation("PAYMENT", "user-123", "/path/to/file.csv")

	// Verify that the error is propagated and the result is nil
	assert.Error(t, err, "Function should return an error when UUID generation fails")
	assert.Nil(t, result, "Result should be nil when an error occurs")
	assert.Contains(t, err.Error(), "simulated UUID generation error", "Error message should be propagated")
}

// TestNewReconciliationWithFixedUUID tests NewReconciliation with a fixed UUID
func TestNewReconciliationWithFixedUUID(t *testing.T) {
	// Save the original generator and restore it after the test
	originalGenerator := defaultUUIDGenerator
	defer func() {
		defaultUUIDGenerator = originalGenerator
	}()

	// Replace the UUID generator with one that returns a fixed UUID
	fixedUUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	defaultUUIDGenerator = func() (uuid.UUID, error) {
		return fixedUUID, nil
	}

	// Call the function
	result, err := NewReconciliation("PAYMENT", "user-123", "/path/to/file.csv")

	// Verify success
	assert.NoError(t, err, "Function should not return an error with successful UUID generation")
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, fixedUUID.String(), result.UUID, "UUID should match the fixed UUID")
}
