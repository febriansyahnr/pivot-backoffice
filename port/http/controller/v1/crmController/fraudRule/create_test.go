package crmfraudrulecontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	validPayload := &fraudrulesmodel.CreateFraudRuleRequest{
		RuleName:      "Test Rule",
		Condition:     "amount > 1000",
		Priority:      1,
		Weight:        decimal.NewFromFloat(0.5),
		IsActive:      true,
		ReferenceType: "transaction",
	}

	createdFraudRule := &fraudrulesmodel.FraudRulesResponse{
		UUID:          "test-uuid",
		RuleName:      validPayload.RuleName,
		Condition:     validPayload.Condition,
		Priority:      validPayload.Priority,
		Weight:        validPayload.Weight,
		IsActive:      validPayload.IsActive,
		ReferenceType: validPayload.ReferenceType,
	}

	expectedSuccess, _ := json.Marshal(map[string]interface{}{
		"code":    "00",
		"message": "Success",
		"data": map[string]interface{}{
			"uuid":          "test-uuid",
			"ruleName":      "Test Rule",
			"condition":     "amount > 1000",
			"priority":      1,
			"weight":        "0.5",
			"isActive":      true,
			"provider":      "",
			"referenceType": "transaction",
			"createdAt":     "0001-01-01T00:00:00Z",
			"updatedAt":     "0001-01-01T00:00:00Z",
		},
	})

	testCases := []struct {
		name         string
		expectedCode int
		expectedBody string
		requestBody  func() []byte
		setup        func(svc *mockSvc.IFraudRuleService)
	}{
		{
			name: "SUCCESS: Create Fraud Rule",
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup: func(svc *mockSvc.IFraudRuleService) {
				svc.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*fraudrulesmodel.CreateFraudRuleRequest"),
				).Return(createdFraudRule, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccess),
		},
		{
			name: "ERROR: Failed to decode request",
			requestBody: func() []byte {
				return []byte("invalid-json")
			},
			setup:        func(svc *mockSvc.IFraudRuleService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid character 'i' looking for beginning of value",
				"error": {
					"type": "API_ERROR",
					"message": "invalid character 'i' looking for beginning of value",
					"recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name: "ERROR: Validation failed",
			requestBody: func() []byte {
				invalidPayload := *validPayload
				invalidPayload.RuleName = ""
				body, _ := json.Marshal(invalidPayload)
				return body
			},
			setup:        func(svc *mockSvc.IFraudRuleService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "Key: 'CreateFraudRuleRequest.RuleName' Error:Field validation for 'RuleName' failed on the 'required' tag",
				"error": {
					"type": "API_ERROR",
					"message": "Key: 'CreateFraudRuleRequest.RuleName' Error:Field validation for 'RuleName' failed on the 'required' tag",
					"recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name: "ERROR: Weight out of range",
			requestBody: func() []byte {
				invalidWeightPayload := *validPayload
				invalidWeightPayload.Weight = decimal.NewFromFloat(2)
				body, _ := json.Marshal(invalidWeightPayload)
				return body
			},
			setup:        func(svc *mockSvc.IFraudRuleService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "weight must be between 0 and 1",
				"error": {
					"type": "API_ERROR",
					"message": "weight must be between 0 and 1",
					"recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name: "ERROR: Create Fraud Rule service error",
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup: func(svc *mockSvc.IFraudRuleService) {
				svc.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*fraudrulesmodel.CreateFraudRuleRequest"),
				).Return(nil, errors.New("internal server error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{
				"code": "99",
				"message": "internal server error",
				"error": {
					"type": "UNKNOWN",
					"message": "internal server error",
					"recommendation": ""
				},
				"data": null
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc := mockSvc.NewIFraudRuleService(t)
			validator := validator.New()

			tc.setup(svc)

			ctrl := New(svc, validator)

			req := httptest.NewRequest(http.MethodPost, "/fraud-rules/create", bytes.NewBuffer(tc.requestBody()))
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.Create)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())
		})
	}
}
