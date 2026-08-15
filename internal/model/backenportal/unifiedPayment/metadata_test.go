package unifiedPaymentModel

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func TestMetadataUnifiedPayment_GetBypassStatusPageState(t *testing.T) {
	tests := []struct {
		name     string
		metadata MetadataUnifiedPayment
		expected bool
	}{
		{
			name: "API mode always returns true regardless of BypassStatusPage value",
			metadata: MetadataUnifiedPayment{
				Mode:             constant.UnifiedPaymentModeAPI,
				BypassStatusPage: false,
			},
			expected: true,
		},
		{
			name: "API mode with BypassStatusPage true returns true",
			metadata: MetadataUnifiedPayment{
				Mode:             constant.UnifiedPaymentModeAPI,
				BypassStatusPage: true,
			},
			expected: true,
		},
		{
			name: "Non-API mode with BypassStatusPage false returns false",
			metadata: MetadataUnifiedPayment{
				Mode:             constant.UnifiedPaymentModeRedirect,
				BypassStatusPage: false,
			},
			expected: false,
		},
		{
			name: "Non-API mode with BypassStatusPage true returns true",
			metadata: MetadataUnifiedPayment{
				Mode:             constant.UnifiedPaymentModeRedirect,
				BypassStatusPage: true,
			},
			expected: true,
		},
		{
			name: "Empty mode with BypassStatusPage false returns false",
			metadata: MetadataUnifiedPayment{
				Mode:             "",
				BypassStatusPage: false,
			},
			expected: false,
		},
		{
			name: "Empty mode with BypassStatusPage true returns true",
			metadata: MetadataUnifiedPayment{
				Mode:             "",
				BypassStatusPage: true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metadata.GetBypassStatusPageState()
			if result != tt.expected {
				t.Errorf("GetBypassStatusPageState() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMetadataUnifiedPaymentIsAutoSplitPaymentAuth(t *testing.T) {
	tests := []struct {
		name     string
		metadata MetadataUnifiedPayment
		expected bool
	}{
		{
			name:     "should return false when AutoSplitPayment is nil",
			metadata: MetadataUnifiedPayment{},
			expected: false,
		},
		{
			name: "should return false when TransactionType is not AUTHENTICATION",
			metadata: MetadataUnifiedPayment{
				AutoSplitPayment: &AutoSplitPayment{
					TransactionType: "CAPTURE",
				},
			},
			expected: false,
		},
		{
			name: "should return true when TransactionType is AUTHENTICATION",
			metadata: MetadataUnifiedPayment{
				AutoSplitPayment: &AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				},
			},
			expected: true,
		},
		{
			name: "should return true with full AutoSplitPayment config",
			metadata: MetadataUnifiedPayment{
				AutoSplitPayment: &AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
					Processor:       "MPGS",
					ProcessorLimit:  2000000000,
					CITMerchantID:   "CIT_MID",
					MITMerchantID:   "MIT_MID",
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metadata.IsAutoSplitPaymentAuth()
			if result != tt.expected {
				t.Errorf("IsAutoSplitPaymentAuth() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
