package constant

import "time"

const (
	RecurringContractTrialTypeFree       = "FREE"
	RecurringContractTrialTypeFixed      = "FIXED"
	RecurringContractTrialTypePercentage = "PERCENTAGE"

	RecurringContractBillingIntervalUnitDay   = "DAY"
	RecurringContractBillingIntervalUnitMonth = "MONTH"
	RecurringContractBillingIntervalUnitYear  = "YEAR"

	RecurringContractStatusCreated         = "CREATED"
	RecurringContractStatusPendInitialAuth = "PENDING_INITIAL_AUTH"
	RecurringContractStatusActive          = "ACTIVE"
	RecurringContractStatusInactive        = "INACTIVE"

	RecurringContractAuthMethodOneDollar    = "ONE_DOLLAR"
	RecurringContractAuthMethodFirstPayment = "FIRST_PAYMENT"

	RecurringContractOneDollarAuthAmountIDR = 10_000.00

	RecurringPaymentMaxProcessDuration = 15 * time.Minute                         // Default value when expiry at value is not specified
	RecurringPaymentMutualExclusionKey = "backend-portal:recurring-payment:%s:%s" // {$1} SUBSEQUENT_PAYMENT/FIRST_AUTHORIZATION, {$2} Recurring ID.

	RecurringPaymentTypeFirstAuthorization = "FIRST_AUTHORIZATION"
	RecurringPaymentTypeSubsequentPayment  = "SUBSEQUENT_PAYMENT"
)

func RecurringMinDaysBetweenPayments(intervalUnit string) uint16 {
	switch intervalUnit {
	case RecurringContractBillingIntervalUnitDay:
		return 1

	case RecurringContractBillingIntervalUnitMonth:
		return 28

	case RecurringContractBillingIntervalUnitYear:
		return 365

	default:
		return 1 // Default acceptable values
	}
}
