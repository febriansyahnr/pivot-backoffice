package unifiedPaymentModel

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type PerformInquiryRequest struct {
	LedgerID string
}

type InquiryDetail struct {
	Status string
}

func (i *InquiryDetail) HasFinalStatus() bool {
	return i.Status == constant.StatusFailed || i.Status == constant.StatusSuccess
}
