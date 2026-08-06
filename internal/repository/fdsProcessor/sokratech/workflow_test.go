package sokratech

import (
	"errors"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/sokratech"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	httpRequestExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	response "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorkflowExecute(t *testing.T) {
	log := loggerMock.NewILogger(t)
	httpClient := httpRequestExtMock.NewIHTTPRequest(t)

	repo := &repository{
		config: config.SokratechConfig{},
		secret: config.SokratechSecret{},
		client: httpClient,
		logger: log,
	}

	workflowID := "ea070a9b-fa16-4cd5-9441-aee5c7b20603"

	tests := []struct {
		name       string
		request    WorkflowRequester
		setupMock  func()
		wantError  error
		wantResult *model.WorkflowExecuteResponse
	}{
		{
			name: "ERROR:Internal error", // NOSONAR
			request: &model.WorkflowExecuteRequest[model.PayoutWorkflowRequest]{
				WorkflowID: "eda8dad2-e1b4-4062-aaa9-6d3bc9bea247",
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil, 0, assert.AnError)
				log.On(
					"Error", mock.Anything, "Failed to request workflow execution", mock.Anything,
				).Once().Return()
				log.On(
					"Info", mock.Anything, "Executing workflow with ID eda8dad2-e1b4-4062-aaa9-6d3bc9bea247", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, assert.AnError),
		},
		{
			name: "ERROR:Partner error", // NOSONAR
			request: &model.WorkflowExecuteRequest[model.PayoutWorkflowRequest]{
				WorkflowID: "13bc9eb5-9985-4507-9f3d-543d12ad3ec1",
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"message":"bad gateway"}`), http.StatusBadGateway, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID 13bc9eb5-9985-4507-9f3d-543d12ad3ec1", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrThirdParty, errors.New(`{"message":"bad gateway"}`)),
		},
		{
			name: "ERROR:Bad request", // NOSONAR
			request: &model.WorkflowExecuteRequest[model.PayoutWorkflowRequest]{
				WorkflowID: workflowID,
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"message":"bad request"}`), http.StatusBadRequest, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrRequest, errors.New(`{"message":"bad request"}`)),
		},
		{
			name: "ERROR:Parse response", // NOSONAR
			request: &model.WorkflowExecuteRequest[model.PayoutWorkflowRequest]{
				WorkflowID: workflowID,
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`Blank`), http.StatusOK, nil)
				log.On(
					"Error", mock.Anything, "Failed to unmarshal workflow response", mock.Anything,
				).Once().Return()
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, constant.ErrInvalidUnmarshalJSON),
		},
		{
			name: "SUCCESS", // NOSONAR
			request: &model.WorkflowExecuteRequest[model.PayoutWorkflowRequest]{
				WorkflowID: workflowID,
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"data":{"result":"approve"}}`), http.StatusOK, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantResult: &model.WorkflowExecuteResponse{
				Data: model.WorkflowExecuteData{
					Result: "approve",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.workflowExecute(t.Context(), test.request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
