package fdscommon

import (
	"encoding/json"
	"time"
)

type AssessPayoutTransactionRequest struct {
	Merchant    Merchant          `json:"merchant"`
	Transaction Transaction       `json:"transaction"`
	Destination PayoutDestination `json:"destination"`
	Metadata    map[string]any    `json:"metadata"`
}

type TransactionAssessmentResponse struct {
	Result              string     `json:"result"`
	RiskScore           int        `json:"riskScore"`
	ExecutionID         string     `json:"executionId,omitempty"`
	WorkflowID          string     `json:"workflowId,omitempty"`
	WorkflowIterationID string     `json:"workflowIterationId,omitempty"`
	WorkflowVersion     string     `json:"workflowVersion,omitempty"`
	ExecutionStartTime  *time.Time `json:"executionStartTime,omitempty"`
}

func (r TransactionAssessmentResponse) RawResult() []byte {
	raw, _ := json.Marshal(r)
	return raw
}
