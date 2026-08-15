package flipProcessorModel

type CommonErrorResponse struct {
	Name    string                `json:"name"`
	Message string                `json:"message"`
	Code    string                `json:"code"`
	Errors  []*FieldErrorResponse `json:"errors,omitempty"`
}

type FieldErrorResponse struct {
	Attribute string `json:"attribute"`
	Code      any    `json:"code"`
	Message   string `json:"message"`
}
