package crmfraudrulecontroller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDetail(t *testing.T) {
	testCases := []struct {
		name         string
		paramID      string
		setup        func(svc *mockSvc.IFraudRuleService)
		expectedCode int
		expectedBody string
	}{
		{
			name:    "SUCCESS: Get Fraud Rule Detail",
			paramID: "123e4567-e89b-12d3-a456-426614174000",
			setup: func(svc *mockSvc.IFraudRuleService) {
				svc.On(
					"Detail",
					mock.Anything,
					"123e4567-e89b-12d3-a456-426614174000",
				).Return(&fraudrulesmodel.FraudRules{
					UUID:          "123e4567-e89b-12d3-a456-426614174000",
					RuleName:      "Test Rule",
					Condition:     "amount > 1000",
					Priority:      1,
					Weight:        decimal.NewFromFloat(0.5),
					IsActive:      true,
					ReferenceType: "transaction",
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"Success","data":{"uuid":"123e4567-e89b-12d3-a456-426614174000","ruleName":"Test Rule","condition":"amount > 1000","priority":1,"weight":"0.5","isActive":true,"provider":"","referenceType":"transaction","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:         "ERROR: Invalid UUID",
			paramID:      "not-a-valid-uuid",
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
			name:    "ERROR: Service Error",
			paramID: "123e4567-e89b-12d3-a456-426614174999",
			setup: func(svc *mockSvc.IFraudRuleService) {
				svc.On(
					"Detail",
					mock.Anything,
					"123e4567-e89b-12d3-a456-426614174999",
				).Return(nil, errors.New("not found"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{
				"code": "99",
				"message": "not found",
				"error": {
					"type": "UNKNOWN",
					"message": "not found",
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

			tc.setup(svc)

			ctrl := New(svc, nil)

			req := httptest.NewRequest(http.MethodGet, "/fraud-rules/"+tc.paramID, nil)
			req = req.WithContext(ctx)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.Detail)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())
		})
	}
}
