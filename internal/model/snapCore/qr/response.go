package snapCoreModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

type GenerateQrMpmResponse struct {
	Data    GenerateQrMpmResponseData `json:"data"`
	Code    string                    `json:"code"`
	Message string                    `json:"message"`
	Error   interface{}               `json:"error,omitempty"`
}

type QueryQrMpmDynamicResponse struct {
	Data    QueryQrMpmDynamicResponseData `json:"data"`
	Code    string                        `json:"code"`
	Message string                        `json:"message"`
	Error   interface{}                   `json:"error,omitempty"`
}

type QueryQrMpmStaticResponse struct {
	Data    QueryQrMpmStaticResponseData `json:"data"`
	Code    string                       `json:"code"`
	Message string                       `json:"message"`
	Error   interface{}                  `json:"error,omitempty"`
}

type GenerateQrMpmResponseData struct {
	UUID               string             `json:"uuid"`
	ResponseCode       string             `json:"responseCode"`
	ResponseMessage    string             `json:"responseMessage"`
	PartnerReferenceNo string             `json:"partnerReferenceNo"`
	ReferenceNo        string             `json:"referenceNo"`
	QrContent          string             `json:"qrContent"`
	MerchantName       string             `json:"merchantName,omitempty"`
	MerchantID         string             `json:"merchantID"`
	StoreID            string             `json:"storeID"`
	QrUrl              string             `json:"qrUrl"`
	ValidityPeriod     *int               `json:"validityPeriod"`
	Amount             commonModel.Amount `json:"amount"`
	FeeAmount          commonModel.Amount `json:"feeAmount,omitempty"`
	QrImage            *string            `json:"qrImage,omitempty"`
	RedirectUrl        *string            `json:"redirectUrl,omitempty"`
	TerminalID         *string            `json:"terminalId,omitempty"`
	Acquirer           string             `json:"acquirer"`
	CreatedAt          time.Time          `json:"createdAt"`
	ExpiredAt          time.Time          `json:"expiredAt"`
	AdditionalInfo     map[string]any     `json:"additionalInfo"`
}

type QueryQrMpmDynamicResponseData struct {
	UUID                string             `json:"uuid"`
	PartnerReferenceNo  string             `json:"partnerReferenceNo"`
	AcquirerReferenceNo string             `json:"acquirerReferenceNo"`
	QrContent           string             `json:"qrContent"`
	MerchantID          string             `json:"merchantID"`
	StoreID             string             `json:"storeID"`
	QrUrl               string             `json:"qrUrl"`
	ValidityPeriod      *int               `json:"validityPeriod"`
	Status              string             `json:"status"`
	QrType              string             `json:"qrType"`
	Amount              commonModel.Amount `json:"amount"`
	RedirectUrl         *string            `json:"redirectUrl,omitempty"`
	TerminalID          *string            `json:"terminalId,omitempty"`
	CreatedAt           time.Time          `json:"createdAt"`
	ExpiredAt           time.Time          `json:"expiredAt"`
	AdditionalInfo      map[string]any     `json:"additionalInfo"`
}

type QueryQrMpmStaticResponseData struct {
	ResponseCode       string                                     `json:"responseCode"`
	ResponseMessage    string                                     `json:"responseMessage"`
	ReferenceNo        string                                     `json:"referenceNo"`
	PartnerReferenceNo string                                     `json:"partnerReferenceNo"`
	AdditionalInfo     TransactionHistoryListPagination           `json:"additionalInfo"`
	DetailData         []TransactionHistoryListResponseDetailData `json:"detailData"`
}

type TransactionHistoryListPagination struct {
	Total   int  `json:"total"`
	Pages   int  `json:"pages"`
	HasMore bool `json:"hasMore"`
}

type TransactionHistoryListResponseDetailData struct {
	DateTime       string                    `json:"dateTime"`
	Amount         commonModel.Amount        `json:"amount"`
	Remark         string                    `json:"remark"`
	OriginalAmount string                    `json:"originalAmount"`
	AdditionalInfo TransactionAdditionalInfo `json:"additionalInfo"`
	Type           string                    `json:"type"`
	Status         string                    `json:"status"`
	SourceOfFunds  *[]SourceOfFunds          `json:"sourceOfFunds,omitempty"`
}

type SourceOfFunds struct {
	Source string             `json:"source"`
	Amount commonModel.Amount `json:"amount"`
}

type TransactionAdditionalInfo struct {
	OriginalPartnerReferenceNo string             `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string             `json:"originalReferenceNo"`
	BankCode                   string             `json:"bankCode"`
	BankName                   string             `json:"bankName"`
	ScID                       string             `json:"scID"`
	Rrn                        string             `json:"rrn"`
	FeeMoney                   commonModel.Amount `json:"feeMoney"`
	Remark                     string             `json:"remark"`
}
