package credential

import "net/http"

type CredentialDashboardReq struct {
	MerchantID string
	UserID     string
	Info       *http.Request
}

type ClientSecretReq struct {
	Action     string        `validate:"-"`
	MerchantID string        `validate:"required,uuid"`
	UserID     string        `validate:"required,uuid"`
	SecretID   string        `validate:"required,uuid"`
	PIN        string        `validate:"-"`
	Info       *http.Request `validate:"-"`
}
