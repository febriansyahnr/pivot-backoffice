package callbackModel

type CallbackRequest struct {
	URL     string
	Request interface{}
}

type CallbackPayloadRequest struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type CallbackSNAPPayloadRequest struct {
	Event  string            `json:"event"`
	Body   interface{}       `json:"body"`
	Header map[string]string `json:"header"`
}
