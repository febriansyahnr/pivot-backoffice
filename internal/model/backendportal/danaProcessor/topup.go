package danaProcessorModel

import commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

type TopupResponse struct {
	ResponseCode       string             `json:"responseCode"`
	ResponseMessage    string             `json:"responseMessage"`
	ReferenceNo        string             `json:"referenceNo"`
	PartnerReferenceNo string             `json:"partnerReferenceNo"`
	SessionId          string             `json:"sessionId"`
	CustomerNumber     string             `json:"customerNumber"`
	Amount             commonModel.Amount `json:"amount"`
	AdditionalInfo     map[string]any     `json:"additionalInfo"`
}

type TopupInquiryAccountResponse struct {
	ResponseCode           string             `json:"responseCode"`
	ResponseMessage        string             `json:"responseMessage"`
	ReferenceNo            string             `json:"referenceNo"`
	PartnerReferenceNo     string             `json:"partnerReferenceNo"`
	SessionId              string             `json:"sessionId"`
	CustomerNumber         string             `json:"customerNumber"`
	CustomerName           string             `json:"customerName"`
	CustomerMonthlyInLimit string             `json:"customerMonthlyInLimit"`
	MinAmount              commonModel.Amount `json:"minAmount"`
	MaxAmount              commonModel.Amount `json:"maxAmount"`
	Amount                 commonModel.Amount `json:"amount"`
	FeeAmount              commonModel.Amount `json:"feeAmount"`
	FeeType                string             `json:"feeType"`
	AdditionalInfo         map[string]any     `json:"additionalInfo"`
}
