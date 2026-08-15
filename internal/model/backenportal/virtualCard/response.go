package virtualCard

type CallbackResponse struct {
	TrxTimestamp       string                   `json:"trxTimestamp"`
	AuthorizationCode  string                   `json:"authorizationCode"`
	ResponseCode       string                   `json:"responseCode"`
	TransactionID      string                   `json:"transactionId"`
	UserReferenceID    string                   `json:"userRefId"`
	CardID             string                   `json:"cardId"`
	CumulativeLimit    float64                  `json:"cumulativeLimit"`
	Amount             float64                  `json:"amount"`
	CurrencyCode       string                   `json:"currencyCode"`
	OriginAmount       float64                  `json:"originAmount"`
	OriginCurrencyCode string                   `json:"originCurrencyCode"`
	Type               string                   `json:"type"`
	Status             string                   `json:"status"`
	ExternalID         string                   `json:"externalId"`
	Merchant           CallbackResponseMerchant `json:"merchant"`
	ResultCode         string                   `json:"resultCode"`
}

type CallbackResponseMerchant struct {
	ID                 string `json:"id"`
	CategoryCode       string `json:"categoryCode"`
	Name               string `json:"name"`
	StateOrCountryCode string `json:"stateOrCountryCode"`
}
