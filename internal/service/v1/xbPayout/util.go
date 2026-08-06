package xbPayoutService

import "github.com/paper-indonesia/pivot-backoffice/constant"

func (s *xbPayoutService) mapStatus(xbStatus string) string {
	switch xbStatus {
	case constant.XbStatusInProcess:
		return constant.XbDisbursementReasonTypePending
	case constant.XbStatusInfoRequested:
		return constant.XbDisbursementReasonTypeDocumentRequested
	case constant.XbStatusComplianceVerification, constant.XbStatusInfoResponded:
		return constant.XbDisbursementReasonTypeInReview
	case constant.XbStatusComplianceRejected:
		return constant.XbDisbursementReasonTypePartnerRejected
	case constant.XbStatusComplianceApproved:
		return constant.XbDisbursementReasonTypeInReview
	case constant.XbStatusPGProcessing:
		return constant.XbDisbursementReasonTypeProcessing
	case constant.XbStatusSentToBank:
		return constant.XbDisbursementReasonTypeProcessing
	case constant.XbStatusRejected:
		return constant.XbDisbursementReasonTypeBeneficiaryRejected
	case constant.XbStatusError:
		// When XB Core Processor returns ERROR status, it means actual failure from provider
		return constant.XbDisbursementReasonTypeFailed
	case constant.XbStatusReturned:
		return constant.XbDisbursementReasonTypeRefunded
	case constant.XbStatusPaid:
		return constant.XbDisbursementReasonTypeSuccess
	case constant.XbStatusAwaitingFunds:
		return constant.XbDisbursementReasonTypePending
	case constant.XbStatusCanceled:
		return constant.XbDisbursementReasonTypeFailed
	case constant.XbStatusRemindRecipient:
		return constant.XbDisbursementReasonTypePending
	default:
		return xbStatus
	}
}
