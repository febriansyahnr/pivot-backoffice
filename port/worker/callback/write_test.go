package callbackWorker_test

import (
	"testing"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/worker/callback"

	conductor "github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWriteCallbackLog(t *testing.T) {

	log := loggerMock.NewILogger(t)
	service := serviceMocks.NewICallbackService(t)

	handler := New(log, service)

	tests := []struct {
		name       string
		task       *conductor.Task
		setupMock  func()
		wantError  error
		wantStatus conductor.TaskResultStatus
		wantOutput map[string]any
	}{
		{
			name: "ERROR:Binding input data", // NOSONAR
			task: &conductor.Task{
				InputData: map[string]any{
					"callbackId": 123456,
				},
			},
			setupMock: func() {
				log.On("Error", mock.Anything, "Failed while binding input data to request model", mock.Anything).Once().Return()
			},
			wantStatus: conductor.FailedWithTerminalErrorTask,
		},
		{
			name: "ERROR:Write callback log", // NOSONAR
			task: &conductor.Task{
				InputData: map[string]any{},
			},
			setupMock: func() {
				service.On("WriteCallbackLogFromWorkflowTask", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed while write callback log from workflow task", mock.Anything).Once().Return()
			},
			wantStatus: conductor.FailedTask,
		},
		{
			name: "SUCCESS", // NOSONAR
			task: &conductor.Task{
				InputData: map[string]any{
					"iteration": 2,
				},
			},
			setupMock: func() {
				service.On("WriteCallbackLogFromWorkflowTask", mock.Anything, mock.Anything).Once().Return(&callbackModel.WorkflowWriteLogResponse{
					CallbackLogId: "f1686fcc-b8d9-4cb4-a80a-ddb8a61d3741", // NOSONAR
				}, nil)
			},
			wantStatus: conductor.CompletedTask,
			wantOutput: map[string]any{
				"callbackLogId": "f1686fcc-b8d9-4cb4-a80a-ddb8a61d3741", // NOSONAR
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			result, err := handler.WriteCallbackLog(t.Context(), test.task)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantStatus, result.Status)
			assert.Equal(t, test.wantOutput, result.OutputData)
		})
	}
}

func TestWriteCallbackMetric(t *testing.T) {

	log := loggerMock.NewILogger(t)

	handler := New(log, nil)

	tests := []struct {
		name       string
		task       *conductor.Task
		setupMock  func()
		wantError  error
		wantStatus conductor.TaskResultStatus
		wantOutput map[string]any
	}{
		{
			name: "ERROR:Binding input data",
			task: &conductor.Task{
				InputData: map[string]any{
					"merchantId": []string{},
				},
			},
			setupMock: func() {
				log.On("Error", mock.Anything, "Failed while binding input data to request model", mock.Anything).Once().Return()
			},
			wantStatus: conductor.FailedWithTerminalErrorTask,
		},
		{
			name: "ERROR:Record metric duration",
			task: &conductor.Task{
				InputData: map[string]any{
					"durationMs": 65,
				},
			},
			setupMock: func() {
				log.On("Warn", mock.Anything, "Failed while sending metrics on workflow write metric", mock.Anything).Times(2).Return()
			},
			wantStatus: conductor.CompletedTask,
		},
		{
			name: "ERROR:Record metric retry count",
			task: &conductor.Task{
				InputData: map[string]any{
					"retryCount": 1,
				},
			},
			setupMock: func() {
				log.On("Warn", mock.Anything, "Failed while sending metrics on workflow write metric", mock.Anything).Once().Return()
			},
			wantStatus: conductor.CompletedTask,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			result, err := handler.WriteCallbackMetric(t.Context(), test.task)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantStatus, result.Status)
			assert.Equal(t, test.wantOutput, result.OutputData)

			log.AssertExpectations(t)
		})
	}
}
