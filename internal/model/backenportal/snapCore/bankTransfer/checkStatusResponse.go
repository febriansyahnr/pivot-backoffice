package banktransfer

import commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

type RequestPayload struct {
	Amount                 commonModel.Amount `json:"amount"`
	BeneficiaryAccountName string             `json:"beneficiaryAccountName"`
	BeneficiaryAccountNo   string             `json:"beneficiaryAccountNo"`
	BeneficiaryAddress     string             `json:"beneficiaryAddress"`
	BeneficiaryBankCode    string             `json:"beneficiaryBankCode"`
	BeneficiaryBankName    string             `json:"beneficiaryBankName"`
	BeneficiaryEmail       string             `json:"beneficiaryEmail"`
	CustomerReference      string             `json:"customerReference"`
	PartnerReferenceNo     string             `json:"partnerReferenceNo"`
	SourceAccountNo        string             `json:"sourceAccountNo"`
	TransactionDate        string             `json:"transactionDate"`
}

type ResponsePayload struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
}

type TransferLog struct {
	UUID            string                     `json:"uuid"`
	Bank            string                     `json:"bank"`
	Action          string                     `json:"action"`
	Status          string                     `json:"status"`
	Order           int                        `json:"order"`
	RequestPayload  RequestPayload             `json:"requestPayload"`
	ResponsePayload ResponsePayload            `json:"responsePayload"`
	AdditionalInfo  *TransferLogAdditionalInfo `json:"additionalInfo"`
	CreatedAt       string                     `json:"createdAt"`
}

type Outbound struct {
	UUID            string                 `json:"uuid"`
	Title           string                 `json:"title"`
	Acquirer        string                 `json:"acquirer"`
	TransactionId   string                 `json:"transactionId"`
	OriginId        string                 `json:"originId"`
	OriginType      string                 `json:"originType"`
	RequestPayload  OutboundRequestPayload `json:"requestPayload"`
	ResponsePayload ResponsePayload        `json:"responsePayload"`
	AdditionalInfo  map[string]any         `json:"additionalInfo"`
	CreatedAt       string                 `json:"createdAt"`
}

type BankTransferCheckStatusResponseData struct {
	UUID         string        `json:"uuid"`
	BankAcquirer string        `json:"bankAcquirer"`
	TransferType string        `json:"transferType"`
	Status       string        `json:"status"`
	TransferLogs []TransferLog `json:"transferLogs"`
	Outbounds    []Outbound    `json:"outbounds"`
}

type BankTransferCheckStatusResponse struct {
	Code    string                              `json:"code"`
	Message string                              `json:"message"`
	Data    BankTransferCheckStatusResponseData `json:"data"`
	Error   interface{}                         `json:"error,omitempty"`
}

type TransferLogAdditionalInfo struct {
	FailedReason *AdditionalInfoFailedReason `json:"failedReason"`
}

type AdditionalInfoFailedReason struct {
	Error string                          `json:"error"`
	Data  *AdditionalInfoFailedReasonData `json:"data"`
}

type AdditionalInfoFailedReasonData struct {
	AdditionalInfo  any    `json:"additionalInfo"`
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
}

type OutboundRequestPayload struct {
	// amount have different type based on acquirer
	Amount                 any    `json:"amount"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo"`
	BeneficiaryAddress     string `json:"beneficiaryAddress"`
	BeneficiaryBankCode    string `json:"beneficiaryBankCode"`
	BeneficiaryBankName    string `json:"beneficiaryBankName"`
	BeneficiaryEmail       string `json:"beneficiaryEmail"`
	CustomerReference      string `json:"customerReference"`
	PartnerReferenceNo     string `json:"partnerReferenceNo"`
	SourceAccountNo        string `json:"sourceAccountNo"`
	TransactionDate        string `json:"transactionDate"`
}
