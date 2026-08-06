package crmfdscontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTransaction(t *testing.T) {
	validUUID := uuid.New().String()
	invalidUUID := "invalid-uuid"

	// Create valid request payload
	validPayload := fdscommon.CRMUpdateRequest{
		AgentCode: "AGENT_001",
		IsFraud:   true,
		Status:    "approved",
		FraudType: "payment risk",
		Note:      "Suspected fraudulent transaction",
		Payment: &fdscommon.CRMPaymentUpdate{
			CardStatus:       "active",
			PaymentStatus:    "paid",
			ChargebackStatus: "opened",
		},
	}

	testCases := []struct {
		desc           string
		setupMocks     func(mockSvc *mocks.IFdsService)
		id             string
		requestBody    interface{}
		expectedStatus int
		expectResponse bool
	}{
		{
			desc: "success - valid request",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, mock.AnythingOfType("*fdscommon.UpdateRequest")).
					Return(&[]fdscommon.UpdateResponse{
						{
							Success: true,
							Code:    util.ValueToPtr("success"),
							Source:  util.ValueToPtr("fraudnet"),
							Message: map[string]string{
								"status": "updated",
							},
						},
					}, nil)
			},
			id:             validUUID,
			requestBody:    validPayload,
			expectedStatus: http.StatusOK,
			expectResponse: true,
		},
		{
			desc:           "error - invalid UUID",
			setupMocks:     func(mockSvc *mocks.IFdsService) {},
			id:             invalidUUID,
			requestBody:    validPayload,
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc:           "error - empty UUID",
			setupMocks:     func(mockSvc *mocks.IFdsService) {},
			id:             "",
			requestBody:    validPayload,
			expectedStatus: http.StatusNotFound, // Chi router returns 404 for empty parameter
			expectResponse: false,
		},
		{
			desc:           "error - invalid JSON body",
			setupMocks:     func(mockSvc *mocks.IFdsService) {},
			id:             validUUID,
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc:           "error - empty request body",
			setupMocks:     func(mockSvc *mocks.IFdsService) {},
			id:             validUUID,
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc:       "error - validation failed - invalid status",
			setupMocks: func(mockSvc *mocks.IFdsService) {},
			id:         validUUID,
			requestBody: fdscommon.CRMUpdateRequest{
				AgentCode: "AGENT_001",
				IsFraud:   true,
				Status:    "invalid_status", // Invalid status
				FraudType: "payment risk",
				Note:      "Test note",
				Payment: &fdscommon.CRMPaymentUpdate{
					CardStatus:       "active",
					PaymentStatus:    "paid",
					ChargebackStatus: "opened",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc:       "error - validation failed - fraud type required when isFraud is true",
			setupMocks: func(mockSvc *mocks.IFdsService) {},
			id:         validUUID,
			requestBody: fdscommon.CRMUpdateRequest{
				AgentCode: "AGENT_001",
				IsFraud:   true,
				Status:    "approved",
				FraudType: "", // Missing fraud type when isFraud is true
				Note:      "Test note",
				Payment: &fdscommon.CRMPaymentUpdate{
					CardStatus:       "active",
					PaymentStatus:    "paid",
					ChargebackStatus: "opened",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc:       "error - validation failed - invalid fraud type",
			setupMocks: func(mockSvc *mocks.IFdsService) {},
			id:         validUUID,
			requestBody: fdscommon.CRMUpdateRequest{
				AgentCode: "AGENT_001",
				IsFraud:   true,
				Status:    "approved",
				FraudType: "invalid_fraud_type", // Invalid fraud type
				Note:      "Test note",
				Payment: &fdscommon.CRMPaymentUpdate{
					CardStatus:       "active",
					PaymentStatus:    "paid",
					ChargebackStatus: "opened",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc:       "error - validation failed - invalid payment status",
			setupMocks: func(mockSvc *mocks.IFdsService) {},
			id:         validUUID,
			requestBody: fdscommon.CRMUpdateRequest{
				AgentCode: "AGENT_001",
				IsFraud:   false,
				Status:    "approved",
				Note:      "Test note",
				Payment: &fdscommon.CRMPaymentUpdate{
					CardStatus:       "active",
					PaymentStatus:    "invalid_payment_status", // Invalid payment status
					ChargebackStatus: "opened",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc: "error - service error",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, mock.AnythingOfType("*fdscommon.UpdateRequest")).
					Return(nil, errors.New("service error"))
			},
			id:             validUUID,
			requestBody:    validPayload,
			expectedStatus: http.StatusBadRequest,
			expectResponse: false,
		},
		{
			desc: "success - minimal valid request (isFraud false)",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, mock.AnythingOfType("*fdscommon.UpdateRequest")).
					Return(&[]fdscommon.UpdateResponse{
						{
							Success: true,
							Code:    util.ValueToPtr("success"),
							Source:  util.ValueToPtr("fraudnet"),
							Message: map[string]string{
								"status": "updated as non-fraud",
							},
						},
					}, nil)
			},
			id: validUUID,
			requestBody: fdscommon.CRMUpdateRequest{
				AgentCode: "AGENT_002",
				IsFraud:   false,
				Status:    "new",
				Note:      "Transaction reviewed",
				Payment: &fdscommon.CRMPaymentUpdate{
					CardStatus:       "active",
					PaymentStatus:    "auth",
					ChargebackStatus: "",
				},
			},
			expectedStatus: http.StatusOK,
			expectResponse: true,
		},
		{
			desc: "success - all valid statuses",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, mock.AnythingOfType("*fdscommon.UpdateRequest")).
					Return(&[]fdscommon.UpdateResponse{
						{
							Success: true,
							Code:    util.ValueToPtr("success"),
							Source:  util.ValueToPtr("fraudnet"),
						},
					}, nil)
			},
			id: validUUID,
			requestBody: fdscommon.CRMUpdateRequest{
				AgentCode: "AGENT_003",
				IsFraud:   true,
				Status:    "cancelled",
				FraudType: "friendly fraud",
				Note:      "Customer dispute",
				Payment: &fdscommon.CRMPaymentUpdate{
					CardStatus:       "suspended",
					PaymentStatus:    "chargeback",
					ChargebackStatus: "lost",
				},
			},
			expectedStatus: http.StatusOK,
			expectResponse: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			config := config.Config{}
			mockSvc := mocks.NewIFdsService(t)
			mockValidator := validator.New()
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMocks(mockSvc)

			controller := New(&config, mockLogger, mockValidator, mockSvc)

			var reqBody *bytes.Buffer
			if tc.requestBody == nil {
				reqBody = bytes.NewBuffer([]byte{})
			} else if str, ok := tc.requestBody.(string); ok {
				reqBody = bytes.NewBufferString(str)
			} else {
				bodyBytes, _ := json.Marshal(tc.requestBody)
				reqBody = bytes.NewBuffer(bodyBytes)
			}

			url := "/crm/v1/fds/update/" + tc.id

			req, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				url,
				reqBody,
			)
			assert.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			testRouter := chi.NewRouter()
			testRouter.Post("/crm/v1/fds/update/{id}", controller.UpdateTransaction)

			testRouter.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectResponse && tc.expectedStatus == http.StatusOK {
				// Verify response contains expected structure
				assert.Contains(t, rr.Body.String(), "success")
				assert.Contains(t, rr.Body.String(), "data")
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestUpdateTransaction_EdgeCases(t *testing.T) {
	validUUID := uuid.New().String()

	testCases := []struct {
		desc           string
		setupRequest   func() *http.Request
		expectedStatus int
	}{
		{
			desc: "error - request with empty body but valid JSON",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/crm/v1/fds/update/"+validUUID,
					strings.NewReader(`{}`), // Empty JSON object
				)
				return req
			},
			expectedStatus: http.StatusBadRequest, // Validation should fail
		},
		{
			desc: "error - request with malformed JSON",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/crm/v1/fds/update/"+validUUID,
					strings.NewReader(`{"agentCode": "AGENT_001", "invalidField"`), // Incomplete JSON
				)
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			config := config.Config{}
			mockSvc := mocks.NewIFdsService(t)
			mockValidator := validator.New()
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Setup mock for cases that might reach the service
			if tc.expectedStatus == http.StatusBadRequest && tc.desc == "error - request with empty body but valid JSON" {
				// This case will pass validation but should fail in service
				mockSvc.On("UpdateTransaction", mock.Anything, validUUID, mock.AnythingOfType("*fdscommon.UpdateRequest")).
					Return(nil, errors.New("validation error"))
			}

			controller := New(&config, mockLogger, mockValidator, mockSvc)

			req := tc.setupRequest()
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			testRouter := chi.NewRouter()
			testRouter.Post("/crm/v1/fds/update/{id}", controller.UpdateTransaction)

			testRouter.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			mockSvc.AssertExpectations(t)
		})
	}
}
