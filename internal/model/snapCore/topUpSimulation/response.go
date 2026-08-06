package snapCoreModel

type TopupSimulationResponse struct {
	Code    string                      `json:"code"`
	Message string                      `json:"message"`
	Data    TopupSimulationResponseData `json:"data"`
	Error   interface{}                 `json:"error,omitempty"`
}

type TopupSimulationResponseData struct {
	VANumber    string `json:"vaNumber"`
	TotalAmount Amount `json:"totalAmount"`
	RequestID   string `json:"requestID"`
	ReferenceNo string `json:"referenceNo"`
}
