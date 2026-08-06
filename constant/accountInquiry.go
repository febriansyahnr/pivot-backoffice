package constant

const (
	RequestAccountInquiryStatusInvalid = "INVALID"
	RequestAccountInquiryStatusWarning = "WARNING"
	RequestAccountInquiryStatusValid   = "VALID"
	RequestAccountInquiryStatusPending = "PENDING"

	ReqInquiryDetailStatusInvalid        = "Account number not found."
	ReqInquiryDetailStatusDormant        = "Account is dormant."
	ReqInquiryDetailStatusInactive       = "Account is inactive."
	ReqInquiryDetailStatusSuspectedFraud = "Suspected fraud account."
	ReqInquiryDetailStatusLimitExceeded  = "Activity count limit exceeded."
	ReqInquiryDetailStatusDoNotHonor     = "Account not honored."
	ReqInquiryDetailStatusPending        = "Inquiry in progress."
	ReqInquiryDetailStatusWarning        = "Incorrect Account name, result from bank: "
)
