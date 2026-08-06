package walletInsights

import "time"

type MerchantTotalBalance struct {
	TotalBalance  float64   `json:"totalBalance"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
}
