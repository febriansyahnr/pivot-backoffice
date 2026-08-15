package fraudnetmodel

type MarketplaceCheckResponse struct {
	Success bool                 `json:"success"`
	Code    *string              `json:"code,omitempty"`
	Source  *string              `json:"source,omitempty"`
	Message any                  `json:"message,omitempty"`
	Data    MarketplaceCheckData `json:"data"`
}

type MarketplaceCheckErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Source  string `json:"source"`
	Message string `json:"message"`
}

type MarketplaceCheckData struct {
	ID        string                 `json:"id"`
	Timer     int                    `json:"timer"`
	RiskScore int                    `json:"risk_score"` // range: 0 - 100
	RiskGroup string                 `json:"risk_group"` // values: very low, low, medium, high, very high
	Link      string                 `json:"link"`
	Tags      []MarketplaceCheckTags `json:"tags"`
}

type MarketplaceCheckTags struct {
	ID        string  `json:"id"`
	Action    *string `json:"action"` // string or null
	Name      string  `json:"name"`
	Source    string  `json:"source"`     // values: rule, label
	Type      string  `json:"type"`       // values: label, queue, workflow
	State     *string `json:"state"`      // nullable
	Weight    *int    `json:"weight"`     // nullable
	RiskScore *int    `json:"risk_score"` // nullable
	RiskGroup *string `json:"risk_group"` // values: very low, low, medium, high, very high
	Link      *string `json:"link"`       // nullable
}
