package sokratech

type WorkflowPayloadRequest interface {
	PayoutWorkflowRequest | PaymentWorkflowRequest
}

type WorkflowExecuteRequest[T WorkflowPayloadRequest] struct {
	WorkflowID   string
	WorkflowName string
	Payload      T
}

func (r *WorkflowExecuteRequest[T]) GetWorkflowID() string {
	return r.WorkflowID
}

func (r *WorkflowExecuteRequest[T]) GetWorkflowName() string {
	return r.WorkflowName
}

func (r *WorkflowExecuteRequest[T]) GetWorkflowPayload() any {
	return r.Payload
}

type PayoutWorkflowRequest struct {
	Merchant    Merchant          `json:"merchant"`
	Transaction Transaction       `json:"transaction"`
	Destination PayoutDestination `json:"destination"`
	Metadata    map[string]any    `json:"metadata"`
}

type PaymentWorkflowRequest struct {
	Merchant      Merchant       `json:"merchant"`
	Customer      Customer       `json:"customer"`
	Transaction   Transaction    `json:"transaction"`
	PaymentMethod PaymentMethod  `json:"paymentMethod"`
	Device        Device         `json:"device"`
	Metadata      map[string]any `json:"metadata"`
}

type ChargebackInfoMetadata struct {
	IsFraud          bool   `json:"isFraud"`
	ChargebackStatus string `json:"chargebackStatus,omitempty"`
	ChargebackNotes  string `json:"chargebackNotes,omitempty"`
}
