package sokratech_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/fdsProcessor/sokratech"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	httpRequestExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	response "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFDSProcessorCheck(t *testing.T) {

	workflowID := "6aa1ffb3-faf3-43d4-975f-9f0bb62eeebd"

	cfg := config.SokratechConfig{
		Workflow: config.SokratechWorkflowConfig{
			PaymentTransactionID: workflowID,
		},
		TimeoutSeconds: 1,
	}

	request := &fdscommon.CheckRequest{
		Partner: fdscommon.PartnerCheck{
			ID:        "merchant-123",               // NOSONAR
			Company:   util.ValueToPtr("Test Corp"), // NOSONAR
			RiskLevel: "MEDIUM",                     // NOSONAR
		},
		Customer: fdscommon.CustomerCheck{
			FirstName: util.ValueToPtr("John"),             // NOSONAR
			Email:     util.ValueToPtr("john@example.com"), // NOSONAR
			Phone:     util.ValueToPtr("+6281234567890"),   // NOSONAR
		},
		Transaction: fdscommon.TransactionCheck{
			OrderID:           "order-456",                                      // NOSONAR
			ClientReferenceID: "client-ref-789",                                 // NOSONAR
			OrderTotal:        util.ValueToPtr(decimal.NewFromFloat(100000.50)), // NOSONAR
			OrderCurrency:     util.ValueToPtr("IDR"),                           // NOSONAR
		},
		Payment: fdscommon.PaymentCheck{
			MethodType:       "CREDIT_CARD",              // NOSONAR
			Fingerprint:      "fp-123",                   // NOSONAR
			MaskedCardNumber: "411111******1111",         // NOSONAR
			CardBrand:        "VISA",                     // NOSONAR
			CardCountryCode:  "ID",                       // NOSONAR
			CardType:         "CREDIT",                   // NOSONAR
			CardIssuing:      "Bank ABC",                 // NOSONAR
			ThreeDsEci:       util.ValueToPtr("05"),      // NOSONAR
			AuthCode:         util.ValueToPtr("AUTH123"), // NOSONAR
			CvvResultCode:    util.ValueToPtr("M"),       // NOSONAR
		},
		Device: fdscommon.DeviceCheck{
			IPType:    "v4",                             // NOSONAR
			IPAddress: util.ValueToPtr("192.168.1.100"), // NOSONAR
		},
		Custom: &fdscommon.CustomCheck{
			Number:        util.ValueToPtr("MID-999"),        // NOSONAR
			AcquiringName: util.ValueToPtr("Acquiring Bank"), // NOSONAR
		},
	}

	log := loggerMock.NewILogger(t)
	httpClient := httpRequestExtMock.NewIHTTPRequest(t)

	repo := New(cfg, config.SokratechSecret{}, httpClient, log).NewFDSProcessor()

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *fdscommon.CheckResponse
	}{
		{
			name: "ERROR:Internal Error", // NOSONAR
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil, 0, assert.AnError)
				log.On(
					"Error", mock.Anything, "Failed to request workflow execution", mock.Anything,
				).Once().Return()
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, assert.AnError),
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
			name: "ERROR:Partner error", // NOSONAR
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"message":"bad gateway"}`), http.StatusBadGateway, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantResult: &fdscommon.CheckResponse{
				Success: false,
			},
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
			wantResult: &fdscommon.CheckResponse{
				Success: true,
				Data: fdscommon.CheckData{
					RiskGroup: "approve",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.Check(t.Context(), request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestFDSProcessorUpdate(t *testing.T) {

	workflowID := "6aa1ffb3-faf3-43d4-975f-9f0bb62eeebd"

	cfg := config.SokratechConfig{
		Workflow: config.SokratechWorkflowConfig{
			PaymentTransactionID: workflowID,
		},
		TimeoutSeconds: 1,
	}

	buildCheckRequest := func() *fdscommon.CheckRequest {
		return &fdscommon.CheckRequest{
			Partner: fdscommon.PartnerCheck{
				ID:        "merchant-123",
				Company:   util.ValueToPtr("Test Corp"),
				RiskLevel: "MEDIUM",
			},
			Customer: fdscommon.CustomerCheck{
				FirstName: util.ValueToPtr("John"),
				Email:     util.ValueToPtr("john@example.com"),
				Phone:     util.ValueToPtr("+6281234567890"),
			},
			Transaction: fdscommon.TransactionCheck{
				OrderID:           "order-456",
				ClientReferenceID: "client-ref-789",
				OrderTotal:        util.ValueToPtr(decimal.NewFromFloat(100000.50)),
				OrderCurrency:     util.ValueToPtr("IDR"),
			},
			Payment: fdscommon.PaymentCheck{
				MethodType: "CREDIT_CARD",
			},
			Device: fdscommon.DeviceCheck{
				IPType:    "v4",
				IPAddress: util.ValueToPtr("192.168.1.100"),
			},
		}
	}

	log := loggerMock.NewILogger(t)
	httpClient := httpRequestExtMock.NewIHTTPRequest(t)

	repo := New(cfg, config.SokratechSecret{}, httpClient, log).NewFDSProcessor()

	tests := []struct {
		name       string
		request    *fdscommon.UpdateRequest
		setupMock  func()
		wantError  error
		wantResult *fdscommon.UpdateResponse
	}{
		{
			name:       "ERROR:Nil request", // NOSONAR
			request:    nil,
			setupMock:  func() {},
			wantResult: &fdscommon.UpdateResponse{Success: false},
		},
		{
			name: "ERROR:Nil FullContext", // NOSONAR
			request: &fdscommon.UpdateRequest{
				FullContext: nil,
			},
			setupMock:  func() {},
			wantResult: &fdscommon.UpdateResponse{Success: false},
		},
		{
			name: "ERROR:Internal Error", // NOSONAR
			request: &fdscommon.UpdateRequest{
				FullContext: buildCheckRequest(),
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil, 0, assert.AnError)
				log.On(
					"Error", mock.Anything, "Failed to request workflow execution", mock.Anything,
				).Once().Return()
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, assert.AnError),
		},
		{
			name: "ERROR:Request timeout", // NOSONAR
			request: &fdscommon.UpdateRequest{
				FullContext: buildCheckRequest(),
			},
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
			name: "ERROR:Partner error", // NOSONAR
			request: &fdscommon.UpdateRequest{
				FullContext: buildCheckRequest(),
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"message":"bad gateway"}`), http.StatusBadGateway, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantResult: &fdscommon.UpdateResponse{
				Success: false,
			},
		},
		{
			name: "SUCCESS:Without IsFraud", // NOSONAR
			request: &fdscommon.UpdateRequest{
				FullContext: buildCheckRequest(),
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"data":{"result":"approve","system_variables":{"execution_id":"exec-001"}}}`), http.StatusOK, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantResult: &fdscommon.UpdateResponse{
				Success: true,
				Data: fdscommon.UpdateData{
					ID: "exec-001",
				},
			},
		},
		{
			name: "SUCCESS:With IsFraud and Payment", // NOSONAR
			request: &fdscommon.UpdateRequest{
				FullContext: buildCheckRequest(),
				IsFraud:     util.ValueToPtr(true),
				Note:        util.ValueToPtr("suspected fraud"),
				Payment: &fdscommon.PaymentUpdate{
					ChargebackStatus: util.ValueToPtr("CHARGEBACK_INITIATED"),
				},
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"data":{"result":"approve","system_variables":{"execution_id":"exec-002"}}}`), http.StatusOK, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantResult: &fdscommon.UpdateResponse{
				Success: true,
				Data: fdscommon.UpdateData{
					ID: "exec-002",
				},
			},
		},
		{
			name: "SUCCESS:With IsFraud without Payment", // NOSONAR
			request: &fdscommon.UpdateRequest{
				FullContext: buildCheckRequest(),
				IsFraud:     util.ValueToPtr(true),
				Note:        util.ValueToPtr("confirmed fraud"),
			},
			setupMock: func() {
				httpClient.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return([]byte(`{"data":{"result":"reject","system_variables":{"execution_id":"exec-003"}}}`), http.StatusOK, nil)
				log.On(
					"Info", mock.Anything, "Executing workflow with ID "+workflowID, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return()
			},
			wantResult: &fdscommon.UpdateResponse{
				Success: true,
				Data: fdscommon.UpdateData{
					ID: "exec-003",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.Update(t.Context(), test.request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
