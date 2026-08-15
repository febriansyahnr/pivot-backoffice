package disbursementModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
)

type ApproveActionObject struct {
	DisbursementID string `json:"id" validate:"uuid"`
}

type RejectActionObject struct {
	DisbursementID    string `json:"id" validate:"uuid"`
	ReasonType        string `json:"reasonType"`
	ReasonDescription string `json:"reasonDescription"`
}

type PayoutObjectForCreate struct {
	ReferenceID        string                   `json:"referenceId" validate:"required"`
	InquiryID          string                   `json:"inquiryId"`
	ChannelCode        string                   `json:"channelCode" validate:"required_without=InquiryID"`
	ChannelInformation PayoutChannelInformation `json:"channelInformation" validate:"required_without=InquiryID"`
	Amount             commonModel.Amount       `json:"amount" validate:"required"`
	Description        string                   `json:"description"`
}

type PayoutObjectForRetry struct {
	ReferenceID        string                   `json:"referenceId"`
	InquiryID          string                   `json:"inquiryId"`
	ChannelCode        string                   `json:"channelCode" validate:"required_without=InquiryID"`
	ChannelInformation PayoutChannelInformation `json:"channelInformation" validate:"required_without=InquiryID"`
	Amount             commonModel.Amount       `json:"amount" validate:"required"`
	Description        string                   `json:"description"`
}

type PayoutResultObject struct {
	TotalPendingCount    int     `json:"totalPendingCount"`
	TotalPendingAmount   float64 `json:"totalPendingAmount"`
	TotalSuccessCount    int     `json:"totalSuccessCount"`
	TotalSuccessAmount   float64 `json:"totalSuccessAmount"`
	TotalFailedCount     int     `json:"totalFailedCount"`
	TotalFailedAmount    float64 `json:"totalFailedAmount"`
	TotalCancelledCount  int     `json:"totalCancelledCount,omitempty"`
	TotalCancelledAmount float64 `json:"totalCancelledAmount,omitempty"`
}

type PayoutObject struct {
	ReferenceID        string                   `json:"referenceId"`
	InquiryID          string                   `json:"inquiryId"`
	ChannelCode        string                   `json:"channelCode"`
	ChannelInformation PayoutChannelInformation `json:"channelInformation"`
	Amount             commonModel.Amount       `json:"amount"`
	Description        string                   `json:"description"`
	Status             string                   `json:"status"`
	Reason             string                   `json:"reason"`
	CreatedAt          time.Time                `json:"created"`
	UpdatedAt          time.Time                `json:"updated"`
}

type PayoutChannelInformation struct {
	AccountNumber string `json:"accountNumber" validate:"omitempty,numeric"`
	AccountName   string `json:"accountName"`
}

type DisbursementReceiptData struct {
	CompletedAt            string `example:"06 May 2024, 04:12:04"`
	ReferenceID            string `example:"ASDou1234"`
	Remark                 string `example:"Payroll May 2024"`
	MerchantName           string `example:"Merchant Name"`
	SenderName             string `example:"PT Harsya Remitindo"`
	Amount                 string `example:"Rp 1.000.000"`
	Status                 string `example:"SUCCESS"`
	DisbursementID         string `example:"uuid-uui-uuid-uuid"`
	BankReferenceNo        string `example:"20240506141204001"`
	BeneficiaryBankName    string `example:"Bank BCA"`
	BeneficiaryAccountNo   string `example:"1234431111"`
	BeneficiaryAccountName string `example:"John Doe"`

	ImageHeader string
}

type OverbookingBankCode struct {
	BankCodes []string `json:"bankCode"`
}

type PayoutCallbackSingleObject struct {
	ReferenceId string              `json:"referenceId"`
	Amount      commonModel.Amount2 `json:"amount"`
	Status      string              `json:"status"`
	Reason      string              `json:"reason,omitempty"`
}
