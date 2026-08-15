package banktransfer

type UpdateTransferStatusPayload struct {
	ExternalID      string `json:"externalId"`
	BankReferenceNo string `json:"bankReferenceNo"`
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	Status          string `json:"status"`
	UUID            string `json:"uuid"`
}
