package sokratech

import "time"

// WorkflowExecuteResponse represents the response from Sokratech workflow execution
type WorkflowExecuteResponse struct {
	Data    WorkflowExecuteData `json:"data"`
	Error   map[string]any      `json:"error"`
	Message string              `json:"message"`
}

// WorkflowExecuteData contains the main data from workflow execution
type WorkflowExecuteData struct {
	Result          string                  `json:"result"`
	Score           int                     `json:"score"`
	Rules           []WorkflowRule          `json:"rules"`
	Transformations map[string]any          `json:"transformations"`
	Aggregations    map[string]any          `json:"aggregations"`
	CustomVariables map[string]any          `json:"custom_variables"`
	HTTPVariables   map[string]any          `json:"http_variables"`
	SystemVariables WorkflowSystemVariables `json:"system_variables"`
	Actions         map[string]any          `json:"actions"`
	Metadata        WorkflowMetadata        `json:"metadata"`
}

// WorkflowRule represents a single rule evaluation result
type WorkflowRule struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Score      int    `json:"score"`
	Result     bool   `json:"result"`
	Snooze     bool   `json:"snooze"`
	Definition string `json:"definition"`
}

// WorkflowSystemVariables contains system-generated variables from workflow execution
type WorkflowSystemVariables struct {
	RiskScore           int       `json:"risk_score"`
	ExecutionID         string    `json:"execution_id"`
	WorkflowID          string    `json:"workflow_id"`
	WorkflowIterationID string    `json:"workflow_iteration_id"`
	WorkflowVersion     string    `json:"workflow_version"`
	ExecutionStartTime  time.Time `json:"execution_start_time"`
}

// WorkflowMetadata contains metadata information about the workflow execution
type WorkflowMetadata struct {
	WorkflowID          string `json:"workflowId"`
	WorkflowIterationID string `json:"workflowIterationId"`
	WorkflowVersion     string `json:"workflowVersion"`
	WorkflowExecutionID string `json:"workflowExecutionId"`
}
