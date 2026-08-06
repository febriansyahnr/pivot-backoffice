package ewallet

import (
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore"
)

type EWalletRefundRequest struct {
	TransactionID string             // Processor Transaction ID
	Amount        commonModel.Amount `json:"amount"`
}

type SnapCoreEWalletRefundResponse struct {
	*snapCoreModel.StandardResponse
	Data *EWalletRefundResponse `json:"data"`
}

type EWalletRefundResponse struct {
	ResponseCode    string             `json:"responseCode"`
	ResponseMessage string             `json:"responseMessage"`
	PartnerRefundNo string             `json:"partnerRefundNo,omitempty"`
	RefundAmount    commonModel.Amount `json:"refundAmount,omitempty"`
	RefundTime      string             `json:"refundTime,omitempty"`
	RefundNo        string             `json:"refundNo,omitempty"`
	AdditionalInfo  map[string]any     `json:"additionalInfo,omitempty"`
}
