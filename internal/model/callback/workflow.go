package callback_model

import (
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/pkg/conductor"

	modelSdk "github.com/conductor-sdk/conductor-go/sdk/model"
)

type WorkflowMerchantCallbackRequest struct {
	Payload string `json:"payload"`
}

type WorkflowMerchantCallbackPreparationResponse struct {
	Name        string  `json:"name"`
	EventName   string  `json:"eventName"`
	MerchantId  string  `json:"merchantId"`
	CallbackId  string  `json:"callbackId"`
	CallbackUrl string  `json:"callbackUrl"`
	Request     string  `json:"request"` // Base64 encoding request body
	IsSnap      bool    `json:"isSnap"`
	ReferenceId *string `json:"referenceId"`
}

type WorkflowRecordMetricRequest struct {
	MerchantId  string `json:"merchantId"`
	EventName   string `json:"eventName"`
	StatusCode  int64  `json:"statusCode"`
	DurationMs  int64  `json:"durationMs"`
	Iteration   int64  `json:"iteration"` // Use this only to count the number of attempts using DO_WHILE loop
	RetryCount  int64  `json:"retryCount"`
	ErrorDetail string `json:"errorDetail"`
}

type WorkflowSendCallbackResponse struct {
	StatusCode     int                                 `json:"statusCode"`
	ResponseBody   map[string]any                      `json:"responseBody,omitempty"`
	AdditionalInfo *SendMerchantCallbackAdditionalInfo `json:"additionalInfo,omitempty"`
	Status         conductor.TaskStatus                `json:"status"`
}

func (w *WorkflowSendCallbackResponse) NonSuccessTaskStatus() modelSdk.TaskResultStatus {
	if w.StatusCode >= 500 || w.StatusCode == 408 || w.StatusCode == 429 {
		return modelSdk.FailedTask
	}
	return modelSdk.FailedWithTerminalErrorTask
}

type WorkflowWriteLogRequest struct {
	CallbackId  string                       `json:"callbackId"`
	EventName   string                       `json:"eventName"`
	IsSnap      bool                         `json:"isSnap"`
	ReferenceId *string                      `json:"referenceId"`
	Payload     string                       `json:"payload"` // Base64 encoding request body
	RawPayload  json.RawMessage              `json:"-"`
	Response    WorkflowSendCallbackResponse `json:"response"`
	RetryCount  int                          `json:"retryCount"`
	Iteration   int                          `json:"iteration"`
	WorkflowId  string                       `json:"workflowId"`
}

type WorkflowWriteLogResponse struct {
	CallbackLogId string `json:"callbackLogId"`
}
