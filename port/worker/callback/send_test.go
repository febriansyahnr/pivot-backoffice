package callbackWorker_test

import (
	"net/http"
	"testing"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/worker/callback"

	conductor "github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendCallback(t *testing.T) {

	log := loggerMock.NewILogger(t)
	service := serviceMocks.NewICallbackService(t)

	handler := New(log, service)

	log.On(
		"Warn", mock.Anything, "Failed while sending metrics on workflow write metric", mock.Anything,
	).Return()

	tests := []struct {
		name                   string
		task                   *conductor.Task
		setupMock              func()
		wantError              error
		wantStatus             conductor.TaskResultStatus
		wantOutput             map[string]any
		wantReasonIncompletion string
	}{
		{
			name: "ERROR:Binding input data", // NOSONAR
			task: &conductor.Task{
				InputData: map[string]any{
					"EventName": []string{},
				},
			},
			setupMock: func() {
				log.On("Error", mock.Anything, "Failed while binding input data to request model", mock.Anything).Once().Return()
			},
			wantStatus:             conductor.FailedWithTerminalErrorTask,
			wantReasonIncompletion: "json: cannot unmarshal array into Go struct field SendMerchantCallbackRequest.eventName of type string",
		},
		{
			name: "ERROR:Send merchant callback", // NOSONAR
			task: &conductor.Task{
				InputData: map[string]any{},
			},
			setupMock: func() {
				service.On("SendMerchantCallback", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatus:             conductor.FailedTask,
			wantReasonIncompletion: "assert.AnError general error for testing",
			wantOutput: map[string]any{
				"statusCode": float64(0),
				"status":     "RETRYABLE_ERROR",
			},
		},
		{
			name: "SUCCESS:Merchant responsed with non-2xx status (non-retryable)", // NOSONAR
			task: &conductor.Task{
				InputData: map[string]any{},
			},
			setupMock: func() {
				service.On("SendMerchantCallback", mock.Anything, mock.Anything).Once().Return(&model.SendMerchantCallbackResponse{
					StatusCode:   http.StatusBadRequest,
					ResponseBody: []byte(`{"message":"bad request"}`),
				}, nil)
			},
			wantStatus:             conductor.FailedWithTerminalErrorTask,
			wantReasonIncompletion: "merchant responded with non-2xx status",
			wantOutput: map[string]any{
				"statusCode":   float64(400),
				"responseBody": map[string]any{"message": "bad request"},
				"status":       "NON_RETRYABLE_ERROR",
			},
		},
		{
			name: "SUCCESS:Merchant responsed with non-2xx status (retryable)", // NOSONAR
			task: &conductor.Task{
				InputData: map[string]any{},
			},
			setupMock: func() {
				service.On("SendMerchantCallback", mock.Anything, mock.Anything).Once().Return(&model.SendMerchantCallbackResponse{
					StatusCode:   http.StatusInternalServerError,
					ResponseBody: []byte(`{"message":"internal error. please try again later."}`),
				}, nil)
			},
			wantStatus:             conductor.FailedTask,
			wantReasonIncompletion: "merchant responded with non-2xx status",
			wantOutput: map[string]any{
				"statusCode":   float64(500),
				"responseBody": map[string]any{"message": "internal error. please try again later."},
				"status":       "RETRYABLE_ERROR",
			},
		},
		{
			name: "SUCCESS:Merchant responsed with non-2xx status", // NOSONAR
			task: &conductor.Task{
				InputData: map[string]any{},
			},
			setupMock: func() {
				service.On("SendMerchantCallback", mock.Anything, mock.Anything).Once().Return(&model.SendMerchantCallbackResponse{
					StatusCode:   http.StatusOK,
					ResponseBody: []byte(`{"message":"Ok"}`),
				}, nil)
			},
			wantStatus: conductor.CompletedTask,
			wantOutput: map[string]any{
				"statusCode":   float64(200),
				"responseBody": map[string]any{"message": "Ok"},
				"status":       "SUCCESS",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			result, err := handler.SendCallback(t.Context(), test.task)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantStatus, result.Status)
			assert.Equal(t, test.wantOutput, result.OutputData)
			assert.Equal(t, test.wantReasonIncompletion, result.ReasonForIncompletion)

			log.AssertExpectations(t)
			service.AssertExpectations(t)
		})
	}
}
