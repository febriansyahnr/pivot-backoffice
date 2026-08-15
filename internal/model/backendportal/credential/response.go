package credential

import "time"

type CredentialDashboardResp struct {
	ClientID      string                `json:"clientId"`
	ClientSecrets []ClientSecretSummary `json:"clientSecrets"`
}

type ClientSecretResp struct {
	Secret     string    `json:"secret"`
	LastUpdate time.Time `json:"lastUpdate"`
	Time       int64     `json:"time"`
}
