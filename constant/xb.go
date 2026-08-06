package constant

import (
	"path/filepath"
	"strings"
)

const (
	// From XB Core Processor
	XbStatusWaiting                = "WAITING"
	XbStatusExpired                = "EXPIRED"
	XbStatusInProcess              = "IN_PROCESS"
	XbStatusAwaitingFunds          = "AWAITING_FUNDS"
	XbStatusInfoRequested          = "INFO_REQUESTED"
	XbStatusInfoResponded          = "INFO_RESPONDED"
	XbStatusComplianceVerification = "COMPLIANCE_VERIFICATION"
	XbStatusComplianceRejected     = "COMPLIANCE_REJECTED"
	XbStatusComplianceApproved     = "COMPLIANCE_APPROVED"
	XbStatusPGProcessing           = "PG_PROCESSING"
	XbStatusSentToBank             = "SENT_TO_BANK"
	XbStatusRejected               = "REJECTED"
	XbStatusError                  = "ERROR"
	XbStatusReturned               = "RETURNED"
	XbStatusPaid                   = "PAID"
	XbStatusPending                = "PENDING"
	XbStatusCanceled               = "CANCELED"
	XbStatusRemindRecipient        = "REMIND_RECIPIENT"
	// From Usecase
	XbStatusCreated   = "CREATED"
	XbStatusConfirmed = "CONFIRMED"
	XbStatusHttpError = "HTTP_ERROR" // Internal status for HTTP confirmation errors (4xx/5xx)
)

// Disbursement Reason Type
const (
	XbDisbursementReasonTypeWaitingForConfirmation = "WAITING_FOR_CONFIRMATION"
	XbDisbursementReasonTypeExpired                = "EXPIRED"
	XbDisbursementReasonTypeInsufficientBalance    = "INSUFFICIENT_BALANCE"
	XbDisbursementReasonTypePending                = "PENDING"
	XbDisbursementReasonTypeDocumentRequested      = "DOCUMENT_REQUESTED"
	XbDisbursementReasonTypeInReview               = "IN_REVIEW"
	XbDisbursementReasonTypePartnerRejected        = "PARTNER_REJECTED"
	XbDisbursementReasonTypeProcessing             = "PROCESSING"
	XbDisbursementReasonTypeBeneficiaryRejected    = "BENEFICIARY_REJECTED"
	XbDisbursementReasonTypeFailed                 = "FAILED"
	XbDisbursementReasonTypeRefunded               = "REFUNDED"
	XbDisbursementReasonTypeSuccess                = "SUCCESS"
	XbDisbursementReasonTypeError                  = "ERROR"
)

// Disbursement Reason Description
const (
	XbDisbursementReasonDescWaitingForConfirmation = "Waiting for merchant to confirm payout"
	XbDisbursementReasonDescExpired                = "Payout expired"
	XbDisbursementReasonDescInsufficientBalance    = "Payout confirmed but merchant’s balance is insufficient"
	XbDisbursementReasonDescPending                = "Payout request sent"
	XbDisbursementReasonDescDocumentRequested      = "Further information (RFI) requested by compliance"
	XbDisbursementReasonDescInReview               = "RFI submitted and in review by compliance"
	XbDisbursementReasonDescPartnerRejected        = "Payout rejected by partner’s compliance"
	XbDisbursementReasonDescProcessing             = "Payout in process and instruction being sent to the bank"
	XbDisbursementReasonDescBeneficiaryRejected    = "Payout rejected by beneficiary through bank"
	XbDisbursementReasonDescFailed                 = "Payout failed"
	XbDisbursementReasonDescRefunded               = "Payout refunded"
	XbDisbursementReasonDescSuccess                = "Payout success and has been received by beneficiary"
	XbDisbursementReasonDescError                  = "Payout failed due to error from provider"
)

// Payout Method
const (
	XbPayoutMethodBank   = "BANK"
	XbPayoutMethodWallet = "WALLET"
	XbPayoutMethodCash   = "CASH"

	DocumentTypeZip = ".zip"
	DocumentTypePdf = ".pdf"
)

// XB Routing Code
const (
	XBRoutingCodeLocal = "LOCAL"
	XBRoutingCodeSwift = "SWIFT"

	XBLocalFee = 10.0 // in USD
	XBSwiftFee = 25.0 // in USD
)

func MapXbProcessorStatusToCoreStatus(status string) (disbursementStatus, disbursementReasonType, accountTransactionStatus string) {
	switch status {
	case XbStatusWaiting:
		return DisbursementStatusWaiting, XbDisbursementReasonTypeWaitingForConfirmation, ""
	case XbStatusExpired:
		return DisbursementStatusRejected, XbDisbursementReasonTypeExpired, ""
	case XbStatusInProcess, XbStatusAwaitingFunds:
		return DisbursementStatusApproved, XbDisbursementReasonTypePending, StatusPending
	case XbStatusInfoRequested:
		return DisbursementStatusApproved, XbDisbursementReasonTypeDocumentRequested, StatusPending
	case XbStatusComplianceVerification:
		return DisbursementStatusApproved, XbDisbursementReasonTypeInReview, StatusPending
	case XbStatusComplianceRejected:
		return DisbursementStatusApproved, XbDisbursementReasonTypePartnerRejected, StatusFailed
	case XbStatusComplianceApproved:
		return DisbursementStatusApproved, XbDisbursementReasonTypeInReview, StatusPending
	case XbStatusPGProcessing, XbStatusSentToBank:
		return DisbursementStatusApproved, XbDisbursementReasonTypeProcessing, StatusPending
	case XbStatusRejected:
		return DisbursementStatusApproved, XbDisbursementReasonTypeBeneficiaryRejected, StatusFailed
	case XbStatusError:
		// When Consumer receives ERROR from XB Core Processor, it means actual failure from provider
		return DisbursementStatusApproved, XbDisbursementReasonTypeFailed, StatusFailed
	case XbStatusReturned:
		return DisbursementStatusApproved, XbDisbursementReasonTypeRefunded, StatusSuccess
	case XbStatusPaid:
		return DisbursementStatusApproved, XbDisbursementReasonTypeSuccess, StatusSuccess
	case XbStatusPending:
		// PENDING from processor because insufficient balance from our side
		// EasyLink only, since NIUM will auto retry
		return DisbursementStatusApproved, XbDisbursementReasonTypePending, StatusPending
	case XbStatusCanceled:
		return DisbursementStatusApproved, XbDisbursementReasonTypeFailed, StatusFailed
	case XbStatusRemindRecipient:
		return DisbursementStatusApproved, XbDisbursementReasonTypePending, StatusPending
	case XbStatusHttpError:
		// HTTP confirmation error (4xx/5xx) - unknown status, retryable
		return DisbursementStatusApproved, XbDisbursementReasonTypeError, StatusPending
	}

	return "", "", ""
}

func MapXbReasonTypeToDesc(reasonType string) string {
	switch reasonType {
	case XbDisbursementReasonTypeWaitingForConfirmation:
		return XbDisbursementReasonDescWaitingForConfirmation
	case XbDisbursementReasonTypeExpired:
		return XbDisbursementReasonDescExpired
	case XbDisbursementReasonTypeInsufficientBalance:
		return XbDisbursementReasonDescInsufficientBalance
	case XbDisbursementReasonTypePending:
		return XbDisbursementReasonDescPending
	case XbDisbursementReasonTypeDocumentRequested:
		return XbDisbursementReasonDescDocumentRequested
	case XbDisbursementReasonTypeInReview:
		return XbDisbursementReasonDescInReview
	case XbDisbursementReasonTypePartnerRejected:
		return XbDisbursementReasonDescPartnerRejected
	case XbDisbursementReasonTypeProcessing:
		return XbDisbursementReasonDescProcessing
	case XbDisbursementReasonTypeBeneficiaryRejected:
		return XbDisbursementReasonDescBeneficiaryRejected
	case XbDisbursementReasonTypeFailed:
		return XbDisbursementReasonDescFailed
	case XbDisbursementReasonTypeRefunded:
		return XbDisbursementReasonDescRefunded
	case XbDisbursementReasonTypeSuccess:
		return XbDisbursementReasonDescSuccess
	case XbDisbursementReasonTypeError:
		return XbDisbursementReasonDescError
	}

	return ""
}

var validDocument = []string{
	DocumentTypeZip,
	DocumentTypePdf,
}

func IsUnderlyingDocumentValidToUpload(filename string) bool {
	if len(filename) < 4 {
		return false
	}

	for _, docType := range validDocument {
		if strings.ToLower(filepath.Ext(filename)) == docType {
			return true
		}
	}

	return false
}

const (
	XBSimulationInsufficientBalanceNumber = "10092025002"
)
