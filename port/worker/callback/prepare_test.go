package callbackWorker_test

import (
	"testing"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	. "github.com/paper-indonesia/pivot-backoffice/port/worker/callback"

	conductor "github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPreparation(t *testing.T) {
	var (
		log     = loggerMock.NewILogger(t)
		service = serviceMocks.NewICallbackService(t)

		payload    = "CgZQQVlPVVQSC1BBWU9VVC5ET05FGiRhZWM2NjM2ZC03YTAyLTRkOTMtYTRjNS0wMDZiOWMyMzUwNjgiogEKOHR5cGUuZ29vZ2xlYXBpcy5jb20vY2FsbGJhY2suRGlzYnVyc2VtZW50Q2FsbGJhY2tSZXF1ZXN0EmYKJGFlYzY2MzZkLTdhMDItNGQ5My1hNGM1LTAwNmI5YzIzNTA2OBIkYWEwODdiYzAtOTRiNC00OTU5LThmNDMtNTZlNDI4NTQwMzQwGhIZAAAAAAAA8D8hAAAAAAD5xUAiBERPTkU="
		merchantId = util.ParseUUID("aec6636d-7a02-4d93-a4c5-006b9c235068")
	)
	outputData := map[string]any{
		"name":        "PAYOUT",
		"eventName":   "PAYOUT.DONE",
		"merchantId":  merchantId.String(),
		"request":     "eyJldmVudCI6IlBBWU9VVC5ET05FIiwiZGF0YSI6eyJtZXJjaGFudElkIjoiYWVjNjYzNmQtN2EwMi00ZDkzLWE0YzUtMDA2YjljMjM1MDY4IiwidXVpZCI6ImFhMDg3YmMwLTk0YjQtNDk1OS04ZjQzLTU2ZTQyODU0MDM0MCIsInBheW91dFJlc3VsdHMiOnsidG90YWxQZW5kaW5nQ291bnQiOjAsInRvdGFsUGVuZGluZ0Ftb3VudCI6MCwidG90YWxTdWNjZXNzQ291bnQiOjEsInRvdGFsU3VjY2Vzc0Ftb3VudCI6MTEyNTAsInRvdGFsRmFpbGVkQ291bnQiOjAsInRvdGFsRmFpbGVkQW1vdW50IjowfSwic3RhdHVzIjoiRE9ORSJ9fQ==",
		"isSnap":      false,
		"callbackId":  "00000000-0000-0000-0000-000000000000",
		"callbackUrl": "",
		"referenceId": "aa087bc0-94b4-4959-8f43-56e428540340",
	}

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
			name: "ERROR:Binding data",
			task: &conductor.Task{
				InputData: map[string]any{"payload": map[string]string{}},
			},
			setupMock: func() {
				log.On("Error", mock.Anything, "Failed while binding input data to request model", mock.Anything).Once().Return()
			},
			wantStatus: conductor.FailedWithTerminalErrorTask,
		},
		{
			name: "ERROR:Decode payload",
			task: &conductor.Task{
				InputData: map[string]any{"payload": "XXXXXXXX{}"},
			},
			setupMock: func() {
				log.On("Error", mock.Anything, "Failed while decode base64 payload to bytes", mock.Anything).Once().Return()
			},
			wantStatus: conductor.FailedWithTerminalErrorTask,
		},
		{
			name: "ERROR:Proto unmarshal",
			task: &conductor.Task{
				InputData: map[string]any{"payload": "MTIzNDU2"},
			},
			setupMock: func() {
				log.On("Error", mock.Anything, "Failed while decompiling proto message", mock.Anything).Once().Return()
			},
			wantStatus: conductor.FailedWithTerminalErrorTask,
		},
		{
			name: "ERROR:Binding request from proto message",
			task: &conductor.Task{
				InputData: map[string]any{"payload": "CgRYWFhYEgRYWFhYGiRiOWZmZTNiMi1mNjRkLTQ0YjgtYTQ2NS0xYzVjYWZhOThiM2YiAA=="},
			},
			setupMock: func() {
				log.On("Error", mock.Anything, "Failed while binding message request to callback request", mock.Anything).Once().Return()
			},
			wantStatus: conductor.FailedWithTerminalErrorTask,
		},
		{
			name: "ERROR:Find callback details",
			task: &conductor.Task{
				InputData: map[string]any{"payload": payload},
			},
			setupMock: func() {
				service.On("FindCallbackByMerchantIdAndCallbackName", mock.Anything, merchantId, "PAYOUT").Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed while find merchant callback details", mock.Anything).Once().Return()
			},
			wantStatus: conductor.FailedTask,
		},
		{
			name: "SUCCESS:Data not found",
			task: &conductor.Task{
				InputData: map[string]any{"payload": payload},
			},
			setupMock: func() {
				service.On("FindCallbackByMerchantIdAndCallbackName", mock.Anything, merchantId, "PAYOUT").Once().Return(nil, nil)
				log.On("Info", mock.Anything, "Data preparation for merchant callback delivery", mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantStatus: conductor.CompletedTask,
			wantOutput: outputData,
		},
		{
			name: "SUCCESS:Data found",
			task: &conductor.Task{
				InputData: map[string]any{"payload": payload},
			},
			setupMock: func() {
				outputData["callbackUrl"] = "http://localhost/events"
				service.On("FindCallbackByMerchantIdAndCallbackName", mock.Anything, merchantId, "PAYOUT").Once().Return(&model.Callback{URL: outputData["callbackUrl"].(string)}, nil)
				log.On("Info", mock.Anything, "Data preparation for merchant callback delivery", mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantStatus: conductor.CompletedTask,
			wantOutput: outputData,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.task == nil {
				test.task = &conductor.Task{}
			}
			test.setupMock()

			result, err := handler.Preparation(t.Context(), test.task)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantStatus, result.Status)
			assert.Equal(t, test.wantOutput, result.OutputData)

			log.AssertExpectations(t)
			service.AssertExpectations(t)
		})
	}
}
