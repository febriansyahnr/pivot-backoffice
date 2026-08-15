package merchant

import "encoding/json"

type ValidateSnapSignatureRequest struct {
	Signature   string          `json:"signature" validate:"required"`
	Timestamp   string          `json:"timestamp" validate:"required,datetime=2006-01-02T15:04:05+07:00"`
	ClientID    string          `json:"clientId" validate:"required"`
	Url         string          `json:"url" validate:"required"`
	Body        json.RawMessage `json:"body" validate:"required"`
	Method      string          `json:"method" validate:"required"`
	AccessToken string          `json:"accessToken"`
}

type GenerateSnapSignatureRequest struct {
	Timestamp   string          `json:"timestamp"`
	ClientID    string          `json:"clientId"`
	Url         string          `json:"url"`
	Body        json.RawMessage `json:"body"`
	Method      string          `json:"method"`
	AccessToken string          `json:"accessToken"`
}

type GenerateSnapB2BTokenSignatureRequest struct {
	Timestamp  string `json:"timestamp" validate:"required,datetime=2006-01-02T15:04:05+07:00"`
	ClientID   string `json:"clientId" validate:"required,uuid"`
	PrivateKey string `json:"privateKey" validate:"required"`
	GrantType  string `json:"grantType" validate:"required"`
}
