package constant

import "net/http"

const (
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
	HeaderContentType   = "Content-Type"

	HeaderXCustomFrom         = "X-Custom-From"
	HeaderXSenderOrigin       = "X-Sender-Origin"
	HeaderXCRMKey             = "X-CRM-Key"
	HeaderXAutomatedTestKey   = "X-Automated-Test-Key"
	HeaderXResponseSignature  = "X-Response-Signature"
	HeaderXRequestPIN         = "X-Request-PIN"
	HeaderXRequestId          = "X-Request-Id"
	HeaderXAPIKey             = "X-API-KEY"
	HeaderXMerchantId         = "X-Merchant-Id"
	HeaderXSubMerchantID      = "X-SubMerchant-Id"
	HeaderXMerchantSecret     = "X-Merchant-Secret"
	HeaderXClientKey          = "X-Client-Key"
	HeaderXClientSecret       = "X-Client-Secret"
	HeaderXInternalServiceKey = "X-Internal-Service-Key"
	HeaderXSimulationKey      = "X-Simulation-Key"
	HeaderXTimestamp          = "X-Timestamp"
	HeaderXEncrypted          = "X-Encrypted"
	HeaderXSignature          = "X-Signature"
	HeaderXPartnerId          = "X-Partner-Id"
	HeaderXExternalId         = "X-External-Id"
	HeaderXChannelId          = "Channel-Id"
	HeaderXSnapPath           = "X-SNAP-Path"
	HeaderXSnapServiceCode    = "X-SNAP-Service-Code"
	HeaderXIdempotencyKey     = "X-Idempotency-Key"
	HeaderXRealIP             = "X-Real-IP"
	HeaderTimezone            = "Time-Zone"
	HeaderXAdvAIKey           = "X-ADVAI-KEY"
	HeaderXEncryptHashKey     = "X-Encrypt-Hash-Key"
	HeaderXPaymentToken       = "X-Payment-Token"
	HeaderXSimulationMode     = "X-Simulation-Mode"
	HeaderXSimulationToken    = "X-Simulation-Token"

	MIMEApplicationJSON = "application/json"

	XForwarderPath = "X-Forwarded-Path"
	// Only for payment simulation in staging environment (Sandbox)
	HeaderXSimulationCardNumber = "X-Simulation-Card-Number"
)

const (
	ActionGet  = http.MethodGet
	ActionPost = http.MethodPost
	ActionPut  = http.MethodPut
)

var IgnoreLoggingPath = []string{
	"/health-check",
	"/ping",
}
