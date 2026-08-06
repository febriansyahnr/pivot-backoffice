package constant

import "slices"

const (
	PaymentMethodChannelTypeAggregator  = "AGGREGATOR"
	PaymentMethodChannelTypeFacilitator = "FACILITATOR" // Deprecated, replaced to DIRECT
	PaymentMethodChannelTypeDirect      = "DIRECT"      // FACILITATOR Replacement

	PaymentMethodActivationStatusNotRequested = "NOT_REQUESTED"
	PaymentMethodActivationStatusRequested    = "REQUESTED"
	PaymentMethodActivationStatusSubmitted    = "SUBMITTED"
	PaymentMethodActivationStatusApproved     = "APPROVED"
	PaymentMethodActivationStatusRejected     = "REJECTED"

	PaymentMethodActivationMethodInstant = "INSTANT"
	PaymentMethodActivationMethodManual  = "MANUAL"
	PaymentMethodActivationMethodApi     = "API"
)

const (
	CacheKeyDefaultVAConfig = "backend-portal:default-va-config:%s:%s"
)

const (
	PaymentMethodGeneralStatusActive   = "ACTIVE"
	PaymentMethodGeneralStatusInactive = "INACTIVE"
)

func IsDirectPSP(s string) bool {
	if slices.Contains([]string{PaymentMethodChannelTypeFacilitator, PaymentMethodChannelTypeDirect}, s) {
		return true
	}

	return false
}
