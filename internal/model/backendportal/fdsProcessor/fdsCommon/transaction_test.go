package fdscommon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransactionAssessmentResponseRawResult(t *testing.T) {
	now := time.Now()

	input := TransactionAssessmentResponse{
		Result:              "approve",
		RiskScore:           25,
		ExecutionID:         "exec-001",
		WorkflowID:          "wf-001",
		WorkflowIterationID: "iter-001",
		WorkflowVersion:     "1.0",
		ExecutionStartTime:  &now,
	}

	result := input.RawResult()

	assert.JSONEq(t, `{
		"result": "approve",
		"riskScore": 25,
		"executionId": "exec-001",
		"workflowId": "wf-001",
		"workflowIterationId": "iter-001",
		"workflowVersion": "1.0",
		"executionStartTime": "`+now.Format(time.RFC3339Nano)+`"
	}`, string(result))
}
