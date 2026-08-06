package crmfraudrulecontroller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-playground/validator/v10"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestList(t *testing.T) {
	expectedSuccess, _ := json.Marshal(map[string]interface{}{
		"code": "00",
		"data": map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"uuid":          "123e4567-e89b-12d3-a456-426614174000",
					"ruleName":      "Rule",
					"condition":     "amount > 1000",
					"priority":      1,
					"weight":        "0.5",
					"isActive":      true,
					"provider":      "",
					"referenceType": "transaction",
					"createdAt":     "0001-01-01T00:00:00Z",
					"updatedAt":     "0001-01-01T00:00:00Z",
				},
			},
			"meta": map[string]interface{}{
				"page":       0,
				"perPage":    0,
				"totalItems": 0,
				"totalPages": 0,
			},
		},
		"message": "Success",
	})

	testCases := []struct {
		name         string
		query        string
		setup        func(svc *mockSvc.IFraudRuleService)
		expectedCode int
		expectedBody string
	}{
		{
			name:  "SUCCESS: List Fraud Rules",
			query: "?ruleName=Rule&referenceType=transaction&page=1&perPage=2",
			setup: func(svc *mockSvc.IFraudRuleService) {
				svc.On("List", mock.Anything, &fraudrulesmodel.FraudRulesQuery{
					RuleName:      "Rule",
					ReferenceType: "transaction",
					Page:          1,
					PageSize:      2,
				}).Return(&commonModel.PaginationResponse{
					Data: []*fraudrulesmodel.FraudRules{
						{
							UUID:          "123e4567-e89b-12d3-a456-426614174000",
							RuleName:      "Rule",
							Condition:     "amount > 1000",
							Priority:      1,
							Weight:        decimal.NewFromFloat(0.5),
							IsActive:      true,
							ReferenceType: "transaction",
						},
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccess),
		},
		{
			name:         "ERROR: Invalid Page Param",
			query:        "?page=zero",
			setup:        func(svc *mockSvc.IFraudRuleService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid page number",
				"error": {
					"type": "API_ERROR",
					"message": "invalid page number", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:         "ERROR: Invalid PerPage Param",
			query:        "?perPage=-1",
			setup:        func(svc *mockSvc.IFraudRuleService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"message": "invalid per page number",
				"error": {
					"type": "API_ERROR",
					"message": "invalid per page number", "recommendation": ""
				},
				"data": null
			}`,
		},
		{
			name:  "ERROR: Service Error",
			query: "?ruleName=fail",
			setup: func(svc *mockSvc.IFraudRuleService) {
				svc.On("List", mock.Anything, mock.Anything).
					Return(nil, errors.New("something went wrong"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{
				"code": "99",
				"message": "something went wrong",
				"error": {
					"type": "UNKNOWN",
					"message": "something went wrong", "recommendation": ""
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

			req := httptest.NewRequest(http.MethodGet, "/fraud-rules"+tc.query, nil)
			req = req.WithContext(ctx)

			// Add raw query
			req.URL.RawQuery = url.Values(req.URL.Query()).Encode()

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.List)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
		})
	}
}
