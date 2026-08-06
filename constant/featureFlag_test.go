package constant_test

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/stretchr/testify/assert"
)

func TestShouldEnforceStandardizedAddress(t *testing.T) {
	// Wait for flag to be loaded
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name              string
		merchantID        string
		merchantCreatedAt time.Time
		expectedResult    bool
		description       string
	}{
		{
			name:              "Merchant in merchantIds array should be enforced",
			merchantID:        "merchant-in-array-id",
			merchantCreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), // Old merchant
			expectedResult:    true,
			description:       "Merchant in merchantIds array should always be enforced regardless of creation date",
		},
		{
			name:              "Old merchant before cutoff date - not enforced",
			merchantID:        "old-merchant-id",
			merchantCreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedResult:    false,
			description:       "Merchant created before cutoff date should not be enforced",
		},
		{
			name:              "Merchant created on cutoff date - enforced",
			merchantID:        "cutoff-merchant-id",
			merchantCreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedResult:    true,
			description:       "Merchant created on cutoff date should be enforced",
		},
		{
			name:              "Merchant created after cutoff date - enforced",
			merchantID:        "new-merchant-id",
			merchantCreatedAt: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			expectedResult:    true,
			description:       "Merchant created after cutoff date should be enforced",
		},
		{
			name:              "Merchant with targeting query - enforced",
			merchantID:        "targeted-merchant-id",
			merchantCreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedResult:    true,
			description:       "Merchant matched by targeting query should be enforced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constant.ShouldEnforceStandardizedAddress(tt.merchantID, tt.merchantCreatedAt)
			assert.Equal(t, tt.expectedResult, result, tt.description)
		})
	}
}

func TestIsPayoutToVirtualAccountAllowed(t *testing.T) {
	tests := []struct {
		bankCode      string
		accountNumber string
		wantAllowed   bool
	}{
		{
			bankCode:      "002",
			accountNumber: "1000000000001",
			wantAllowed:   true,
		},
		{
			bankCode:      "014",
			accountNumber: "123450000001",
			wantAllowed:   true,
		},
		{
			bankCode:      "014",
			accountNumber: "543211234001",
			wantAllowed:   true,
		},
		{
			bankCode:      "002",
			accountNumber: "123450000001",
			wantAllowed:   false,
		},
		{
			bankCode:      "002",
			accountNumber: "543211234001",
			wantAllowed:   false,
		},
		{
			bankCode:      "002",
			accountNumber: "543201234001",
			wantAllowed:   true,
		},
		{
			bankCode:      "002",
			accountNumber: "54322234001",
			wantAllowed:   true,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.wantAllowed, constant.IsPayoutToVirtualAccountAllowed(tt.bankCode, tt.accountNumber))
	}
}
