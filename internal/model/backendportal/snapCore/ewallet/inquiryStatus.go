package ewallet

import (
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

type EWalletInquiryStatusRequest struct {
	TransactionID string // Processor Transaction ID
}

type EWalletInquiryStatusResponse struct {
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

type InquiryStatusAdditionalInfo struct {
	AmountDetail AmountDetail           `json:"amountDetail"`
	Buyer        Buyer                  `json:"buyer"`
	CloseReason  string                 `json:"closeReason"`
	Seller       map[string]interface{} `json:"seller"`
	StatusDetail StatusDetail           `json:"statusDetail"`
	TimeDetail   TimeDetail             `json:"timeDetail"`
}

type AmountDetail struct {
	ChargeAmount     commonModel.Amount `json:"chargeAmount"`
	ChargeBackAmount commonModel.Amount `json:"chargeBackAmount"`
	ConfirmAmount    commonModel.Amount `json:"confirmAmount"`
	OrderAmount      commonModel.Amount `json:"orderAmount"`
	PayAmount        commonModel.Amount `json:"payAmount"`
	RefundAmount     commonModel.Amount `json:"refundAmount"`
	VoidAmount       commonModel.Amount `json:"voidAmount"`
}

type Buyer struct {
	UserID string `json:"userId"`
}

type StatusDetail struct {
	AcquirementStatus string `json:"acquirementStatus"` // CLOSED
	Frozen            string `json:"frozen"`            // TRUE / FALSE
}

type TimeDetail struct {
	CreatedTime string   `json:"createdTime"`
	ExpiryTime  string   `json:"expiryTime"`
	PaidTimes   []string `json:"paidTimes"`
}

type PaymentViewDetail struct {
	PaidTime        string          `json:"paidTime"`
	PayOptionsInfos []PayOptionInfo `json:"payOptionInfos"`
}

type PayOptionInfo struct {
	ChargeAmount commonModel.Amount `json:"chargeAmount"`
	PayAmount    commonModel.Amount `json:"payAmount"`
	TransAmount  commonModel.Amount `json:"transAmount"`
	PayMethod    string             `json:"payMethod"`
}
