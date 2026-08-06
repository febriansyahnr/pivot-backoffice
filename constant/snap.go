package constant

type TReconCode string

func (t TReconCode) Message() string {
	switch t {
	case ReconCodeInvalidAmount:
		return "Invalid amount"
	case ReconCodeInvalidReference:
		return "Invalid reference"
	case ReconCodeInvalidStatus:
		return "Invalid status"
	case ReconCodeIvalidDate:
		return "Invalid date"
	case ReconCodeOk:
		return "OK"
	default:
		return "Unknown"
	}
}

const (
	ReconSnapStatusInvalid    = "INVALID"
	ReconSnapStatusValid      = "VALID"
	ReconCodeInvalidAmount    = TReconCode("INVALID_AMOUNT")
	ReconCodeInvalidReference = TReconCode("INVALID_REFERENCE")
	ReconCodeInvalidStatus    = TReconCode("INVALID_STATUS")
	ReconCodeIvalidDate       = TReconCode("INVALID_DATE")
	ReconCodeOk               = TReconCode("OK")
	ReconTimeFormat           = "January 2, 2006 3:04 PM"
)

const (
	GetVirtualAccountSnapApiCode          = "30"
	CreateVirtualAccountSnapApiCode       = "27"
	UpdateVirtualAccountSnapApiCode       = "28"
	GenerateQrisMPMSnapApiCode            = "47"
	PaymentNotifyQrisMPMSnapApiCode       = "52"
	QueryPaymentDynamicQrisMPMSnapApiCode = "51"
	QueryPaymentStaticQrisMPMSnapApiCode  = "12"
)

const (
	QrLatestStatusSuccess  = "SUCCESS"
	QrLatestStatusRefunded = "REFUNDED"
	QrLatestStatusCanceled = "CANCELED"
	QrLatestStatusFailed   = "FAILED"
)

var (
	SnapLatestTransactionStatusSuccess    = "00"
	SnapLatestTransactionStatusProcessing = []string{"01", "02", "07"}
	SnapLatestTransactionStatusCancelled  = "05"
	SnapLatestTransactionStatusFailed     = "06"
	SnapLatestTransactionStatusNotFound   = "07"
)

var (
	UnauthorizedSnapFmt               = "Unauthorized %s"
	InvalidB2BTokenSnapErrMsg         = "Invalid Token (B2B)"
	InvalidMandatoryFieldSnapFmt      = "Invalid Mandatory Field %s"
	InvalidFieldFormatSnapFmt         = "Invalid Field Format %s"
	TransactionNotFoundSnapErrMsg     = "Transaction Not Found"
	InvalidAmountSnapErrMsg           = "Invalid Amount"
	DuplicatePartnerReferenceNoErrMsg = "Duplicate partnerReferenceNo"
	ConflictErrMsg                    = "Conflict"
	InvalidBillVirtualAccountErrMsg   = "Invalid Bill/Virtual Account"
	PartnerNotFoundErrMsg             = "Partner Not Found"
	BadRequest                        = "Bad Request"
	InvalidMerchant                   = "Invalid Merchant"
)
