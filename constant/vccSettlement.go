package constant

import "time"

const (
	VccSettlementBilledStatus = "BILLED"

	BillingCycleFirst = 1
	BillingCycleLast  = 12
)

const (
	ProcessRcnTransactionInquiryCacheLockKey  = "%s:vcc:settlement:transaction:inquiry:%s:%s:%s:lock"  // %s:vcc:settlement:transaction:inquiry:{merchantId}:{rcnId}:{postingDate}:lock
	ProcessRcnTransactionInquiryCacheStateKey = "%s:vcc:settlement:transaction:inquiry:%s:%s:%s:state" // %s:vcc:settlement:transaction:inquiry:{merchantId}:{rcnId}:{postingDate}:state

	ProcessRcnTransactionInquiryCacheLockDuration  = time.Minute * 30
	ProcessRcnTransactionInquiryCacheStateDuration = time.Minute * 30
)
