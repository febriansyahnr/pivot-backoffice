package merchant

import "encoding/base64"

type JITApiKeyResponse struct {
	APIKey string `json:"apiKey"`
}

func ToJITAPIKeyResponse(jitApiKey string) *JITApiKeyResponse {
	return &JITApiKeyResponse{
		APIKey: base64.StdEncoding.EncodeToString([]byte(jitApiKey)),
	}
}
