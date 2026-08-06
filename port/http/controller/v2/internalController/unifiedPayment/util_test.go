package v2InternalUnifiedPaymentController

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestValidateAmountRange(t *testing.T) {
	controller := &paymentController{}

	tests := []struct {
		name      string
		amount    float64
		minAmount *float64
		maxAmount *float64
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid amount within range",
			amount:    5000.0,
			minAmount: util.ValueToPtr(1000.0),
			maxAmount: util.ValueToPtr(10000.0),
			wantErr:   false,
		},
		{
			name:      "Valid amount at minimum",
			amount:    1000.0,
			minAmount: util.ValueToPtr(1000.0),
			maxAmount: util.ValueToPtr(10000.0),
			wantErr:   false,
		},
		{
			name:      "Valid amount at maximum",
			amount:    10000.0,
			minAmount: util.ValueToPtr(1000.0),
			maxAmount: util.ValueToPtr(10000.0),
			wantErr:   false,
		},
		{
			name:      "Valid amount with no min/max constraints",
			amount:    999999.0,
			minAmount: nil,
			maxAmount: nil,
			wantErr:   false,
		},
		{
			name:      "Valid amount with only min constraint",
			amount:    5000.0,
			minAmount: util.ValueToPtr(1000.0),
			maxAmount: nil,
			wantErr:   false,
		},
		{
			name:      "Valid amount with only max constraint",
			amount:    5000.0,
			minAmount: nil,
			maxAmount: util.ValueToPtr(10000.0),
			wantErr:   false,
		},
		{
			name:      "ERROR: Amount below minimum",
			amount:    500.0,
			minAmount: util.ValueToPtr(1000.0),
			maxAmount: util.ValueToPtr(10000.0),
			wantErr:   true,
			errMsg:    constant.ErrPaymentBelowMinAmount.Error(),
		},
		{
			name:      "ERROR: Amount above maximum",
			amount:    15000.0,
			minAmount: util.ValueToPtr(1000.0),
			maxAmount: util.ValueToPtr(10000.0),
			wantErr:   true,
			errMsg:    constant.ErrPaymentAboveMaxAmount.Error(),
		},
		{
			name:      "ERROR: Amount below minimum (no max)",
			amount:    500.0,
			minAmount: util.ValueToPtr(1000.0),
			maxAmount: nil,
			wantErr:   true,
			errMsg:    constant.ErrPaymentBelowMinAmount.Error(),
		},
		{
			name:      "ERROR: Amount above maximum (no min)",
			amount:    15000.0,
			minAmount: nil,
			maxAmount: util.ValueToPtr(10000.0),
			wantErr:   true,
			errMsg:    constant.ErrPaymentAboveMaxAmount.Error(),
		},
		{
			name:      "Edge case: Zero amount with min constraint",
			amount:    0.0,
			minAmount: util.ValueToPtr(1000.0),
			maxAmount: util.ValueToPtr(10000.0),
			wantErr:   true,
			errMsg:    constant.ErrPaymentBelowMinAmount.Error(),
		},
		{
			name:      "Edge case: Zero min and max",
			amount:    100.0,
			minAmount: util.ValueToPtr(0.0),
			maxAmount: util.ValueToPtr(0.0),
			wantErr:   true,
			errMsg:    constant.ErrPaymentAboveMaxAmount.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := controller.validateAmountRange(tt.amount, tt.minAmount, tt.maxAmount)

			if tt.wantErr {
				assert.Error(t, err)
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, err)
			}
		})
	}
}

func TestIsCybersourceTestAmountAllowed(t *testing.T) {
	// NOTE: This test relies on feature flags being initialized in TestMain (type_test.go)
	// The configuration is in test/consul/backend-portal/feature-flag.yaml

	tests := []struct {
		name        string
		environment string
		merchantID  string
		amount      float64
		expected    bool
		description string
	}{
		{
			name:        "Should return false when merchant ID is empty",
			environment: constant.EnvironmentStaging,
			merchantID:  "",
			amount:      100.0,
			expected:    false,
			description: "Empty merchant ID should fail validation",
		},
		{
			name:        "Should return false when environment is not staging",
			environment: constant.EnvironmentProduction,
			merchantID:  "whitelisted-merchant-id",
			amount:      100.0,
			expected:    false,
			description: "Production environment should not allow test amounts",
		},
		{
			name:        "Should return false when environment is development",
			environment: constant.EnvironmentDevelopment,
			merchantID:  "whitelisted-merchant-id",
			amount:      100.0,
			expected:    false,
			description: "Development environment should not allow test amounts",
		},
		{
			name:        "Should return false when environment is local",
			environment: constant.EnvironmentLocal,
			merchantID:  "whitelisted-merchant-id",
			amount:      100.0,
			expected:    false,
			description: "Local environment should not allow test amounts",
		},
		{
			name:        "Should return TRUE for whitelisted merchant with whitelisted amount",
			environment: constant.EnvironmentStaging,
			merchantID:  "whitelisted-merchant-id",
			amount:      100.0,
			expected:    true,
			description: "Staging environment with whitelisted merchant and amount should pass",
		},
		{
			name:        "Should return TRUE for second merchant with whitelisted amount",
			environment: constant.EnvironmentStaging,
			merchantID:  "922e39ab-7565-49f6-b84f-fb56122821ae",
			amount:      50.0,
			expected:    true,
			description: "Second merchant with whitelisted amount should pass",
		},
		{
			name:        "Should return TRUE for whitelisted merchant (amount filtering works in Consul)",
			environment: constant.EnvironmentStaging,
			merchantID:  "whitelisted-merchant-id",
			amount:      999.0,
			expected:    true,
			description: "Feature flag matches merchant_id condition (amount filtering requires Consul)",
		},
		{
			name:        "Should return FALSE for non-whitelisted merchant",
			environment: constant.EnvironmentStaging,
			merchantID:  "non-whitelisted-merchant",
			amount:      100.0,
			expected:    false,
			description: "Non-whitelisted merchant should fail even with valid amount",
		},
		{
			name:        "Should return false for staging with empty merchant ID",
			environment: constant.EnvironmentStaging,
			merchantID:  "",
			amount:      1000.0,
			expected:    false,
			description: "Even in staging, empty merchant ID should fail",
		},
		{
			name:        "Edge case: Zero amount in staging",
			environment: constant.EnvironmentStaging,
			merchantID:  "whitelisted-merchant-id",
			amount:      0.0,
			expected:    true,
			description: "Feature flag matches merchant_id (amount validation in Consul)",
		},
		{
			name:        "Edge case: Negative amount in staging",
			environment: constant.EnvironmentStaging,
			merchantID:  "whitelisted-merchant-id",
			amount:      -100.0,
			expected:    true,
			description: "Feature flag matches merchant_id (amount validation in Consul)",
		},
		{
			name:        "Edge case: All whitelisted amounts for first merchant",
			environment: constant.EnvironmentStaging,
			merchantID:  "whitelisted-merchant-id",
			amount:      500.0,
			expected:    true,
			description: "500 is whitelisted for first merchant",
		},
		{
			name:        "Edge case: All whitelisted amounts for first merchant (1000)",
			environment: constant.EnvironmentStaging,
			merchantID:  "whitelisted-merchant-id",
			amount:      1000.0,
			expected:    true,
			description: "1000 is whitelisted for first merchant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Environment: tt.environment,
			}
			controller := &paymentController{
				config: cfg,
			}

			payload := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				MerchantID: tt.merchantID,
				Amount: unifiedPaymentModel.Amount{
					Value: tt.amount,
				},
			}

			result := controller.isCybersourceTestAmountAllowed(payload.MerchantID, payload.Amount.Value)

			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}
