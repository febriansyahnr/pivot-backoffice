package snapCoreModel

type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	TraceId string `json:"traceId"`
}

type StandardResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

type SnapURLParam struct {
	URL        string `json:"url"`
	IsDeepLink string `json:"isDeepLink"`
}
