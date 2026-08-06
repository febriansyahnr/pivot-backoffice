package cimbProcessorModel

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
	XSignature    string `json:"X-SIGNATURE"`
}
