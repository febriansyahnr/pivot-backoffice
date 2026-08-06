package settlementHold

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSettlementHold(t *testing.T) {
	now := time.Now().UTC()
	settlementHoldID := uuid.NewString()
	merchantID := uuid.NewString()
	paymentID := uuid.NewString()

	validPayload := &settlementHold.CreateUpdateSettlementHoldRequest{
		MerchantID: merchantID,
		PaymentID:  paymentID,
		Action:     "HOLD",
		Reason:     "Suspicious transaction",
		CreatedBy:  "admin@paper.id",
	}

	createdSettlementHold := &settlementHold.CreateUpdateSettlementHoldResponse{
		UUID:       settlementHoldID,
		MerchantID: merchantID,
		PaymentID:  paymentID,
		Status:     "HOLD",
		Reason:     "Suspicious transaction",
		CreatedBy:  "admin@paper.id",
		CreatedAt:  now,
		UpdatedBy:  "admin@paper.id",
		UpdatedAt:  now,
	}

	expectedSuccess, _ := json.Marshal(map[string]interface{}{
		"code":    "00",
		"message": "OK",
		"data": map[string]interface{}{
			"uuid":       settlementHoldID,
			"merchantId": merchantID,
			"paymentId":  paymentID,
			"status":     "HOLD",
			"reason":     "Suspicious transaction",
			"createdBy":  "admin@paper.id",
			"createdAt":  now.Format(time.RFC3339Nano),
			"updatedBy":  "admin@paper.id",
			"updatedAt":  now.Format(time.RFC3339Nano),
		},
	})

	testCases := []struct {
		name         string
		expectedCode int
		expectedBody string
		requestBody  func() []byte
		setup        func(svc *mockSvc.ISettlementHoldService)
	}{
		{
			name: "SUCCESS: Create settlement hold",
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup: func(svc *mockSvc.ISettlementHoldService) {
				svc.On(
					"CreateUpdate",
					mock.Anything,
					mock.AnythingOfType("*settlementHold.CreateUpdateSettlementHoldRequest"),
				).Return(createdSettlementHold, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccess),
		},
		{
			name: "ERROR: Failed to decode request",
			requestBody: func() []byte {
				return []byte("invalid-json")
			},
			setup:        func(svc *mockSvc.ISettlementHoldService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid character 'i' looking for beginning of value",
				"error": {
					"type": "API_ERROR",
					"details": [],
					"traceId": ""
				},
				"data": null
			}`,
		},
		{
			name: "ERROR: Validation failed - missing paymentId",
			requestBody: func() []byte {
				invalidPayload := *validPayload
				invalidPayload.PaymentID = ""
				body, _ := json.Marshal(invalidPayload)
				return body
			},
			setup:        func(svc *mockSvc.ISettlementHoldService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Validation failed - missing action",
			requestBody: func() []byte {
				invalidPayload := *validPayload
				invalidPayload.Action = ""
				body, _ := json.Marshal(invalidPayload)
				return body
			},
			setup:        func(svc *mockSvc.ISettlementHoldService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Validation failed - invalid action",
			requestBody: func() []byte {
				invalidPayload := *validPayload
				invalidPayload.Action = "INVALID"
				body, _ := json.Marshal(invalidPayload)
				return body
			},
			setup:        func(svc *mockSvc.ISettlementHoldService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Validation failed - missing reason",
			requestBody: func() []byte {
				invalidPayload := *validPayload
				invalidPayload.Reason = ""
				body, _ := json.Marshal(invalidPayload)
				return body
			},
			setup:        func(svc *mockSvc.ISettlementHoldService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Validation failed - missing createdBy",
			requestBody: func() []byte {
				invalidPayload := *validPayload
				invalidPayload.CreatedBy = ""
				body, _ := json.Marshal(invalidPayload)
				return body
			},
			setup:        func(svc *mockSvc.ISettlementHoldService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "ERROR: Service error",
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup: func(svc *mockSvc.ISettlementHoldService) {
				svc.On(
					"CreateUpdate",
					mock.Anything,
					mock.AnythingOfType("*settlementHold.CreateUpdateSettlementHoldRequest"),
				).Return(nil, errors.New("internal server error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{
				"code": "99",
				"message": "internal server error",
				"error": {
					"type": "UNKNOWN",
					"details": [],
					"traceId": ""
				},
				"data": null
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc := mockSvc.NewISettlementHoldService(t)
			validator := validator.New()

			tc.setup(svc)

			ctrl := New(svc, validator)

			req := httptest.NewRequest(http.MethodPost, "/settlement-hold/create", bytes.NewBuffer(tc.requestBody()))
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.CreateSettlementHold)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			if tc.expectedBody != "" {
				assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())
			}
		})
	}
}
