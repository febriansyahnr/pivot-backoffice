package util

import "github.com/paper-indonesia/pivot-backoffice/constant"

func MapAccountInquirySnapResponseToDetailStatus(responseCode string) string {
	switch {
	case IsPatternMatch(constant.SnapCoreResponseCodeDormantAccountPattern, responseCode):
		return constant.ReqInquiryDetailStatusDormant
	case IsPatternMatch(constant.SnapCoreResponseCodeInactiveAccountPattern, responseCode):
		return constant.ReqInquiryDetailStatusInactive
	case IsPatternMatch(constant.SnapCoreSuspectedFraudCodePattern, responseCode):
		return constant.ReqInquiryDetailStatusSuspectedFraud
	case IsPatternMatch(constant.SnapCoreActivityCountLimitExceededCodePattern, responseCode):
		return constant.ReqInquiryDetailStatusLimitExceeded
	case IsPatternMatch(constant.SnapCoreDoNotHonorCodePattern, responseCode):
		return constant.ReqInquiryDetailStatusDoNotHonor
	case IsPatternMatch(constant.SnapCoreResponseCodeInvalidAccountPattern, responseCode):
		return constant.ReqInquiryDetailStatusInvalid
	default:
		return constant.ReqInquiryDetailStatusInvalid
	}
}

func MapQRLatestStatusToPaymentStatus(qrStatus string) string {
	switch qrStatus {
	default:
		return constant.StatusPending
	case constant.QrLatestStatusSuccess, constant.QrLatestStatusRefunded:
		return constant.StatusSuccess
	case constant.QrLatestStatusCanceled, constant.QrLatestStatusFailed:
		return constant.StatusFailed
	}

}
