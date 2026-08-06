package constant

const (
	MerchantFeeAmountType           = "AMOUNT"
	MerchantFeePercentageType       = "PERCENTAGE"
	MerchantFeeAmountPercentageType = "AMOUNT_PERCENTAGE"

	MerchantTaxTypeInclusive = "INCLUSIVE"
	MerchantTaxTypeExclusive = "EXCLUSIVE"
	MerchantTaxTypeNonPKP    = "NON_PKP"

	MerchantFeeDeductionTypeDirect    = "DIRECT"
	MerchantFeeDeductionTypeAutomated = "AUTOMATED"
	MerchantFeeDeductionTypeManual    = "MANUAL"

	TPVTieringType       = "TPV"
	FrequencyTieringType = "FREQUENCY"

	MonthlyAssessedTieringModel = "MONTHLY_ASSESSED"
	LadderTieringModel          = "LADDER"
)

const (
	FeeOnBehalfTypeNotSet  = "NOT_SET"
	FeeOnBehalfTypeAll     = "ALL"
	FeeOnBehalfTypeDefault = "DEFAULT"
	FeeOnBehalfTypeDirect  = "DIRECT"
)

const (
	MerchantFeeInstallmentChannelFormat = "%s_%dM" // BRI_1M, BRI_3M
)
