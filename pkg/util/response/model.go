package response

// Existing Response, will be deleted later
type Response struct {
	Code  string      `json:"code,omitempty" example:"00"`
	Error interface{} `json:"errors,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Meta  interface{} `json:"meta,omitempty"`
}

// Internal Response for OpenAPI
type OpenApiResponse struct {
	Code       string        `json:"code" example:"00"`
	Message    string        `json:"message"`
	Error      *OpenApiError `json:"error,omitempty"`
	Data       interface{}   `json:"data"`
	Pagination interface{}   `json:"pagination,omitempty"`
}

type OpenApiError struct {
	Type           string `json:"type"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
}

// API Response for Merchant Portal
type ApiResponse struct {
	Code       string      `json:"code" example:"00"`
	Message    string      `json:"message"`
	Error      *ApiError   `json:"error,omitempty"`
	Data       interface{} `json:"data"`
	Pagination interface{} `json:"pagination,omitempty"`
}

type ApiError struct {
	Type    string           `json:"type"`
	Source  string           `json:"source,omitempty"`
	Details []ApiErrorDetail `json:"details"`
	TraceId string           `json:"traceId"`
}

type ApiErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type OpenApiSnapResp struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
}

type OpenApiErrorNonSnap struct {
	Code       string      `json:"code" example:"00"`
	Message    string      `json:"message"`
	Error      *ApiError   `json:"error,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
}
