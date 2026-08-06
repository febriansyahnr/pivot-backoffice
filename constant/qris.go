package constant

const (
	QrMethodTypeMPM = "MPM"
	QrMethodTypeCPM = "CPM"

	QrTypeDynamic = "DYNAMIC"
	QrTypeStatic  = "STATIC"

	QrStatusActive   = "ACTIVE"
	QrStatusInactive = "INACTIVE"

	QrisDynamicValidityPeriodMax = 28000
	SnapQrisTypeDynamicMinAmount = 10000
	SnapQrisTypeDynamicMaxAmount = 10000000
)

// QRIS Registration Constant
const (
	PositionDirector     = "Director"
	PositionCommissioner = "Commissioner"

	EnterpriseType = "Enterprise"

	FillingFormReg = "FILLING_FORM"
	SubmittedReg   = "SUBMITTED"
	SuccessReg     = "SUCCESS"
	FailedReg      = "FAILED"

	QrMerchantTypeMerchant    = "Merchant"
	QrMerchantTypeSubMerchant = "Sub-Merchant"
	QrMerchantTypeFranchisee  = "Franchisee"

	QrRegistrationStatusSuccess = "SUCCESS"
)
