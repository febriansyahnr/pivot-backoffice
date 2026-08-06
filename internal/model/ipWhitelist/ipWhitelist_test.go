package ipwhitelistModel

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	IPAddress = "1.1.1.1"
	Subnet    = "24"
)

func TestNew(t *testing.T) {
	merchantId := uuid.NewString()

	// Save the original generator and restore it after the test
	originalGenerator := defaultUUIDGenerator
	defer func() {
		defaultUUIDGenerator = originalGenerator
	}()

	testCases := []struct {
		name     string
		input    *CreateIPWhitelistConfiguration
		expected *IPWhitelistConfiguration
		wantErr  bool
	}{
		{
			name: "SUCCESS: With subnet",
			input: &CreateIPWhitelistConfiguration{
				IP:          IPAddress,
				Subnet:      Subnet,
				MerchantID:  merchantId,
				Priority:    0,
				Action:      "ALLOW",
				Description: "description",
			},
			expected: &IPWhitelistConfiguration{
				ID:          uuid.NewString(),
				MerchantID:  merchantId,
				IP:          IPAddress,
				Subnet:      Subnet,
				Priority:    0,
				Action:      "ALLOW",
				Status:      "ACTIVE",
				Description: "description",
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			},
		},
		{
			name: "SUCCESS: Without subnet",
			input: &CreateIPWhitelistConfiguration{
				IP:          IPAddress,
				Subnet:      "",
				MerchantID:  merchantId,
				Priority:    0,
				Action:      "ALLOW",
				Description: "description",
			},
			expected: &IPWhitelistConfiguration{
				ID:          uuid.NewString(),
				MerchantID:  merchantId,
				IP:          IPAddress,
				Subnet:      "",
				Priority:    0,
				Action:      "ALLOW",
				Status:      "ACTIVE",
				Description: "description",
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			},
		},
		{
			name: "ERROR: Invalid IP",
			input: &CreateIPWhitelistConfiguration{
				IP:          "1.1.1.",
				Subnet:      Subnet,
				MerchantID:  merchantId,
				Priority:    0,
				Action:      "ALLOW",
				Description: "description",
			},
			expected: &IPWhitelistConfiguration{
				ID:          uuid.NewString(),
				MerchantID:  merchantId,
				IP:          IPAddress,
				Subnet:      Subnet,
				Priority:    0,
				Action:      "ALLOW",
				Status:      "ACTIVE",
				Description: "description",
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := New(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err, "Expected an error but got nil")
				assert.Nil(t, result, "Result should be nil when there's an error")

				// Verify that the error is related to IP validation
				assert.Contains(t, err.Error(), "invalid", "Error should indicate invalid IP")
			} else {
				assert.Nil(t, err, "Expected no error but got: %v", err)
				assert.NotNil(t, result, "Result should not be nil when there's no error")
				assert.Equal(t, tc.expected.MerchantID, result.MerchantID)
				assert.Equal(t, tc.expected.IP, result.IP)
				assert.Equal(t, tc.expected.Subnet, result.Subnet)
				assert.Equal(t, tc.expected.Priority, result.Priority)
				assert.Equal(t, tc.expected.Action, result.Action)
				assert.Equal(t, tc.expected.Status, result.Status)
				assert.Equal(t, tc.expected.Description, result.Description)

				// Additional assertions to verify the generated fields
				assert.NotEmpty(t, result.ID, "ID should not be empty")
				_, err := uuid.Parse(result.ID)
				assert.NoError(t, err, "ID should be a valid UUID")

				assert.True(t, result.CreatedAt.Before(time.Now().UTC().Add(time.Second)), "CreatedAt should be in the past")
				assert.True(t, result.UpdatedAt.Before(time.Now().UTC().Add(time.Second)), "UpdatedAt should be in the past")
			}
		})
	}
}

// TestNewErrorHandling tests the error handling of New when UUID generation fails
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
	merchantId := uuid.NewString()
	request := &CreateIPWhitelistConfiguration{
		IP:          IPAddress,
		Subnet:      Subnet,
		MerchantID:  merchantId,
		Priority:    0,
		Action:      "ALLOW",
		Description: "description",
	}

	// Call the function with the mock generator
	result, err := New(request)

	// Verify that the error is propagated and the result is nil
	assert.Error(t, err, "Function should return an error when UUID generation fails")
	assert.Nil(t, result, "Result should be nil when an error occurs")
	assert.Contains(t, err.Error(), "simulated UUID generation error", "Error message should be propagated")
}

func TestUpdate(t *testing.T) {
	merchantId := uuid.NewString()

	testCases := []struct {
		name    string
		input   *UpdateIPWhitelistConfiguration
		base    *IPWhitelistConfiguration
		wantErr bool
	}{
		{
			name: "SUCCESS: With subnet",
			input: &UpdateIPWhitelistConfiguration{
				IP:          IPAddress,
				Subnet:      Subnet,
				MerchantID:  merchantId,
				Priority:    0,
				Action:      "BLOCK",
				Description: "description",
				Status:      "ACTIVE",
			},
			base: &IPWhitelistConfiguration{
				ID:          uuid.NewString(),
				MerchantID:  merchantId,
				IP:          IPAddress,
				Subnet:      Subnet,
				Priority:    0,
				Action:      "ALLOW",
				Status:      "ACTIVE",
				Description: "description",
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			},
		},
		{
			name: "SUCCESS: Without subnet",
			input: &UpdateIPWhitelistConfiguration{
				IP:          IPAddress,
				Subnet:      "",
				MerchantID:  merchantId,
				Priority:    0,
				Action:      "ALLOW",
				Description: "description",
				Status:      "ACTIVE",
			},
			base: &IPWhitelistConfiguration{
				ID:          uuid.NewString(),
				MerchantID:  merchantId,
				IP:          IPAddress,
				Subnet:      "",
				Priority:    0,
				Action:      "ALLOW",
				Status:      "ACTIVE",
				Description: "description",
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			},
		},
		{
			name: "ERROR: Invalid IP",
			input: &UpdateIPWhitelistConfiguration{
				IP:          "1.1.1.",
				Subnet:      Subnet,
				MerchantID:  merchantId,
				Priority:    0,
				Action:      "ALLOW",
				Description: "description",
				Status:      "ACTIVE",
			},
			base: &IPWhitelistConfiguration{
				ID:          uuid.NewString(),
				MerchantID:  merchantId,
				IP:          IPAddress,
				Subnet:      Subnet,
				Priority:    0,
				Action:      "ALLOW",
				Status:      "ACTIVE",
				Description: "description",
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save the original UpdatedAt time for comparison
			originalUpdatedAt := tc.base.UpdatedAt
			originalIP := tc.base.IP
			originalSubnet := tc.base.Subnet
			originalAction := tc.base.Action
			originalStatus := tc.base.Status
			originalDescription := tc.base.Description

			// Wait a small amount of time to ensure UpdatedAt will be different
			time.Sleep(1 * time.Millisecond)

			err := tc.base.Update(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err, "Expected an error but got nil")

				// Verify that the error is related to IP validation
				assert.Contains(t, err.Error(), "invalid", "Error should indicate invalid IP")

				// Verify that the object was not modified when an error occurred
				assert.Equal(t, originalIP, tc.base.IP, "IP should not change when error occurs")
				assert.Equal(t, originalSubnet, tc.base.Subnet, "Subnet should not change when error occurs")
				assert.Equal(t, originalAction, tc.base.Action, "Action should not change when error occurs")
				assert.Equal(t, originalStatus, tc.base.Status, "Status should not change when error occurs")
				assert.Equal(t, originalDescription, tc.base.Description, "Description should not change when error occurs")
				assert.Equal(t, originalUpdatedAt, tc.base.UpdatedAt, "UpdatedAt should not change when error occurs")
			} else {
				assert.Nil(t, err, "Expected no error but got: %v", err)
				assert.Equal(t, tc.base.MerchantID, tc.input.MerchantID)
				assert.Equal(t, tc.base.IP, tc.input.IP)
				assert.Equal(t, tc.base.Subnet, tc.input.Subnet)
				assert.Equal(t, tc.base.Priority, tc.input.Priority)
				assert.Equal(t, tc.base.Action, tc.input.Action)
				assert.Equal(t, tc.base.Status, tc.input.Status)
				assert.Equal(t, tc.base.Description, tc.input.Description)

				// Verify that UpdatedAt was changed
				assert.True(t, tc.base.UpdatedAt.After(originalUpdatedAt),
					"UpdatedAt should be updated to a later time")
			}
		})
	}
}

func TestToResponseModel(t *testing.T) {
	merchantId := uuid.NewString()
	config := &IPWhitelistConfiguration{
		ID:          uuid.NewString(),
		MerchantID:  merchantId,
		IP:          IPAddress,
		Subnet:      Subnet,
		Priority:    0,
		Action:      "BLOCK",
		Status:      "ACTIVE",
		Description: "description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	response := config.ToResponseModel()
	assert.Equal(t, config.ID, response.ID)
	assert.Equal(t, config.MerchantID, response.MerchantID)
	assert.Equal(t, config.IP, response.IP)
	assert.Equal(t, config.Subnet, response.Subnet)
	assert.Equal(t, config.Priority, response.Priority)
	assert.Equal(t, config.Action, response.Action)
	assert.Equal(t, config.Status, response.Status)
	assert.Equal(t, config.Description, response.Description)
	assert.Equal(t, config.CreatedAt, response.CreatedAt)
	assert.Equal(t, config.UpdatedAt, response.UpdatedAt)
}
