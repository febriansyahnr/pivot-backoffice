package snapCoreModel

type BankCodeListResponse struct {
	Data    BankCodeListResponseData `json:"data"`
	Code    string                   `json:"code"`
	Message string                   `json:"message"`
	Error   interface{}              `json:"error,omitempty"`
}

type BankCodeListResponseData struct {
	TransferType string    `json:"transferType" example:"INTRABANK"`
	BankCodes    *[]string `json:"bankCodes"`
}
