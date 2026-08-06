package sokratech_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/fdsProcessor/sokratech"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	httpRequestExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	response "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAssessPayoutTransaction(t *testing.T) {

	workflowID := "16dc5368-9e6d-4bd0-8dea-84501b5cc919"

	cfg := config.SokratechConfig{
		Workflow: config.SokratechWorkflowConfig{
			PayoutTransactionID: workflowID,
		},
		TimeoutSeconds: 1,
	}

	log := loggerMock.NewILogger(t)
	httpClient := httpRequestExtMock.NewIHTTPRequest(t)

	repo := New(cfg, config.SokratechSecret{}, httpClient, log)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *fdscommon.TransactionAssessmentResponse
	}{
		{
			name: "ERROR:Partner error", // NOSONAR
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"message":"bad gateway"}`), http.StatusBadGateway, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrThirdParty, errors.New(`{"message":"bad gateway"}`)),
		},
		{
			name: "ERROR:Request timeout", // NOSONAR
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Run(func(args mock.Arguments) {

					time.Sleep(1_020 * time.Millisecond)
					require.Equal(t, context.DeadlineExceeded, args.Get(0).(context.Context).Err())

				}).Return(nil, 0, context.DeadlineExceeded)
				log.On(
					"Error", mock.Anything, "Failed to request workflow execution", mock.Anything,
				).Once().Return()

				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, context.DeadlineExceeded),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"data":{"result":"approve"}}`), http.StatusOK, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantResult: &fdscommon.TransactionAssessmentResponse{
				Result:             constant.WorkflowFDSResultApprove, // NOSONAR
				ExecutionStartTime: &time.Time{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.AssessPayoutTransaction(t.Context(), fdscommon.AssessPayoutTransactionRequest{})
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
