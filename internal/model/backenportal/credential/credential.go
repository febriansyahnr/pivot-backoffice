package credential

import "time"

type CredentialDashboard struct {
	ClientID      string                `db:"client_id"`
	ClientSecrets []ClientSecretSummary `db:"-"`
}

type ClientSecretSummary struct {
	ID         string    `json:"id" db:"id"`
	KeyName    string    `json:"keyName" db:"key_name"`
	LastUpdate time.Time `json:"lastUpdate" db:"updated_at"`
}

func (d *CredentialDashboard) ToResponse() *CredentialDashboardResp {
	return &CredentialDashboardResp{
		ClientID:      d.ClientID,
		ClientSecrets: d.ClientSecrets,
	}
}
