package fraudnetmodel

type MarketplaceUpdateResponse struct {
	Success bool                  `json:"success"`
	Code    *string               `json:"code,omitempty"`
	Source  *string               `json:"source,omitempty"`
	Message any                   `json:"message,omitempty"`
	Data    MarketplaceUpdateData `json:"data"`
}

type MarketplaceUpdateData struct {
	ID    string `json:"id"`
	Link  string `json:"link"`
	Timer int    `json:"timer"`
}
