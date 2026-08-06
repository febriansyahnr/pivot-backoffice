package danaProcessorModel

type SnapGenericResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
}

type B2bTokenResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	AccessToken     string `json:"accessToken"`
	TokenType       string `json:"tokenType"`
	ExpiresIn       string `json:"expiresIn"`
}

type Header struct {
	ContentType   string `json:"Content-Type"`
	Authorization string `json:"Authorization"`
	XTimestamp    string `json:"X-TIMESTAMP"`
	XPartnerID    string `json:"X-PARTNER-ID"`
	XExternalID   string `json:"X-EXTERNAL-ID"`
	ChannelID     string `json:"CHANNEL-ID"`
	XSignature    string `json:"X-SIGNATURE"`
	XAPIKey       string `json:"X-API-KEY"`
}
