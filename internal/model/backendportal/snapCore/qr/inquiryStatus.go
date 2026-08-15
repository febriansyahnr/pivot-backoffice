package snapCoreModel

type QrisInquiryStatusResponse struct {
	Code    string                         `json:"code"`
	Message string                         `json:"message"`
	Data    *QrisInquiryStatusResponseData `json:"data,omitempty"`
	Error   any                            `json:"error,omitempty"`
}

type QrisInquiryStatusResponseData struct {
	ResponseCode        string               `json:"responseCode"`
	ResponseMessage     string               `json:"responseMessage"`
	UUID                string               `json:"uuid"`
	TransactionID       string               `json:"transactionId"`
	PartnerReferenceNo  string               `json:"partnerReferenceNo"`
	AcquirerReferenceNo string               `json:"acquirerReferenceNo"`
	Status              string               `json:"status"`
	QrType              string               `json:"qrType"`
	Acquirer            string               `json:"acquirer"`
	Amount              *InquiryStatusAmount `json:"amount,omitempty"`
	AdditionalInfo      map[string]any       `json:"additionalInfo,omitempty"`
}

type InquiryStatusAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}
