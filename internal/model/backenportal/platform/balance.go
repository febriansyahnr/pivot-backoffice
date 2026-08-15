package platform

type GetBulkBalanceRequest struct {
	MerchantID string
	Usecase    string `json:"usecase" validate:"required,oneof=DISBURSEMENT PAYMENT"`
	Page       int    `json:"page" validate:"omitempty,min=1"`
	PerPage    int    `json:"perPage" validate:"omitempty,min=1,max=30"`
}

type MerchantBalanceResponse struct {
	MerchantID       string                            `json:"merchantId"`
	AvailableBalance *PlatformAvailableBalanceResponse `json:"availableBalance"`
}

type PlatformAvailableBalanceResponse struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}
