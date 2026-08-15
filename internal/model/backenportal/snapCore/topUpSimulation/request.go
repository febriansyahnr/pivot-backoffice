package snapCoreModel

type TopupSimulationRequest struct {
	MerchantId  string `json:"merchantId"`
	AccountName string `json:"accountName"`
	VANumber    string `json:"vaNumber" validate:"required"`
	TotalAmount Amount `json:"totalAmount"`
}

type Amount struct {
	Value    string `json:"value" validate:"required,min=1"`
	Currency string `json:"currency" validate:"required,oneof=IDR"`
}
