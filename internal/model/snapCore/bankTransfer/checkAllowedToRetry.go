package snapCoreModel

type CheckAllowedToRetryRequest struct {
	Force      bool   `json:"force"`
	ExternalID string `json:"externalId"`
	MerchantId string `json:"-"`
}

type CheckAllowedToRetryResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}
