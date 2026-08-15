package ewallet

import (
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore"
)

type EWalletPaymentSimulationRequest struct {
	Acquirer            string             `json:"acquirer" validate:"required"` // SHOPEEPAY, DANA, etc
	OriginalReferenceID string             `json:"originalReferenceId" validate:"required"`
	Status              string             `json:"status" validate:"required,oneof=SUCCESS FAILED"`
	Amount              commonModel.Amount `json:"amount" validate:"required"`
}

type SnapCoreEWalletPaymentSimulationResponse struct {
	*snapCoreModel.StandardResponse
	Data *EWalletPaymentSimulationResponse `json:"data"`
}

type EWalletPaymentSimulationResponse struct {
	ResponseCode                 string             `json:"responseCode"`
	ResponseMessage              string             `json:"responseMessage"`
	LatestTransactionStatus      string             `json:"latestTransactionStatus"`
	OriginalReferenceNo          string             `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo   string             `json:"originalPartnerReferenceNo"`
	ServiceCode                  string             `json:"serviceCode"`
	TransactionStatusDescription string             `json:"transactionStatusDesc"`
	Title                        string             `json:"title"`
	Amount                       commonModel.Amount `json:"amount"`
	TransAmount                  commonModel.Amount `json:"trans_amount"`
	AdditionalInfo               map[string]any     `json:"additionalInfo"`
}
