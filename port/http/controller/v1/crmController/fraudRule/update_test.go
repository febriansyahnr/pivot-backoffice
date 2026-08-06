package crmfraudrulecontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	validID := "123e4567-e89b-12d3-a456-426614174000"
	validPayload := fraudrulesmodel.UpdateFraudRuleRequest{
		RuleName:      util.ValueToPtr("Updated Rule"),
		Condition:     util.ValueToPtr("amount > 2000"),
		Priority:      util.ValueToPtr(2),
		Weight:        util.ValueToPtr(decimal.NewFromFloat(0.7)),
		IsActive:      util.ValueToPtr(true),
		ReferenceType: util.ValueToPtr("transaction"),
	}

	expectedResponse := &fraudrulesmodel.FraudRulesResponse{
		UUID:          validID,
		RuleName:      *validPayload.RuleName,
		Condition:     *validPayload.Condition,
		Priority:      *validPayload.Priority,
		Weight:        *validPayload.Weight,
		IsActive:      *validPayload.IsActive,
		ReferenceType: *validPayload.ReferenceType,
	}

	expectedSuccess, _ := json.Marshal(map[string]interface{}{
		"code":    "00",
		"message": "Success",
		"data": map[string]interface{}{
			"uuid":          validID,
			"ruleName":      validPayload.RuleName,
			"condition":     validPayload.Condition,
			"priority":      validPayload.Priority,
			"weight":        "0.7",
			"isActive":      true,
			"provider":      "",
			"referenceType": validPayload.ReferenceType,
			"createdAt":     "0001-01-01T00:00:00Z",
			"updatedAt":     "0001-01-01T00:00:00Z",
		},
	})

	testCases := []struct {
		name         string
		paramID      string
		requestBody  func() []byte
		setup        func(svc *mockSvc.IFraudRuleService)
		expectedCode int
		expectedBody string
	}{
		{
			name:    "SUCCESS: Update Fraud Rule",
			paramID: validID,
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup: func(svc *mockSvc.IFraudRuleService) {
				expected := validPayload
				expected.UUID = validID
				svc.On("Update", mock.Anything, &expected).Return(expectedResponse, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccess),
		},
		{
			name:         "ERROR: Invalid UUID",
			paramID:      "invalid-uuid",
			requestBody:  func() []byte { return []byte("{}") },
			setup:        func(svc *mockSvc.IFraudRuleService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid id",
				"error": {
					"type": "API_ERROR",
					"message": "invalid id",
					"recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Malformed JSON body",
			paramID:      validID,
			requestBody:  func() []byte { return []byte("invalid-json") },
			setup:        func(svc *mockSvc.IFraudRuleService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"errors": "invalid character 'i' looking for beginning of value"
			}`,
		},
		{
			name:    "ERROR: Update service failure",
			paramID: validID,
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup: func(svc *mockSvc.IFraudRuleService) {
				expected := validPayload
				expected.UUID = validID
				svc.On("Update", mock.Anything, &expected).Return(nil, errors.New("internal failure"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"99","message":"internal failure","error":{"type":"UNKNOWN","message":"internal failure","recommendation":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc := mockSvc.NewIFraudRuleService(t)
			validator := validator.New()

			tc.setup(svc)

			ctrl := New(svc, validator)

			req := httptest.NewRequest(http.MethodPut, "/fraud-rules/"+tc.paramID, bytes.NewBuffer(tc.requestBody()))
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.Update)
			handler.ServeHTTP(rr, req)

			t.Logf("Handler response body: %s", rr.Body.String())

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())
		})
	}
}
