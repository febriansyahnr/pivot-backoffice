package callback_model_test

import (
	"net/http"
	"testing"

	modelSdk "github.com/conductor-sdk/conductor-go/sdk/model"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/callback"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowSendCallbackResponse(t *testing.T) {
	tests := []struct {
		statusCode     int
		wantTaskStatus modelSdk.TaskResultStatus
	}{
		{
			statusCode:     http.StatusInternalServerError,
			wantTaskStatus: modelSdk.FailedTask,
		},
		{
			statusCode:     408,
			wantTaskStatus: modelSdk.FailedTask,
		},
		{
			statusCode:     http.StatusTooManyRequests,
			wantTaskStatus: modelSdk.FailedTask,
		},
		{
			statusCode:     http.StatusBadRequest,
			wantTaskStatus: modelSdk.FailedWithTerminalErrorTask,
		},
		{
			statusCode:     http.StatusNotFound,
			wantTaskStatus: modelSdk.FailedWithTerminalErrorTask,
		},
	}
	for _, test := range tests {
		response := WorkflowSendCallbackResponse{
			StatusCode: test.statusCode,
		}
		assert.Equal(t, test.wantTaskStatus, response.NonSuccessTaskStatus())
	}
}
