package installmentplan

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
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	merchantID := uuid.New().String()
	midID := uuid.New().String()

	validPayload := &installmentPlanModel.UpdateInstallmentPlanRequest{
		UUID:           "test-uuid-123",
		MerchantID:     merchantID,
		Acquirer:       "HARSYA",
		SettlementType: "AGGREGATOR",
		PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
		Title:          "Test Plan",
		Description:    "Test Description",
		CardDetail: &installmentPlanModel.UpdateCardInstallmentPlanRequest{
			MidID:       midID,
			AllowedBins: []string{"123456", "654321"},
		},
	}

	createdInstallmentPlan := &installmentPlanModel.InstallmentPlan{
		UUID:           "test-uuid-123",
		MerchantID:     merchantID,
		Acquirer:       "HARSYA",
		SettlementType: "AGGREGATOR",
		PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
		Title:          "Test Plan",
		Description:    "Test Description",
		Tenor:          12,
		Status:         constant.InstallmentPlanStatusActive,
	}

	expectedSuccess, _ := json.Marshal(map[string]interface{}{
		"code":    "00",
		"message": "OK",
		"data": map[string]interface{}{
			"uuid":            "test-uuid-123",
			"merchantId":      merchantID,
			"acquirer":        "HARSYA",
			"settlementType":  "AGGREGATOR",
			"installmentType": "",
			"paymentMethod":   constant.InstallmentPlanPaymentMethodCard,
			"title":           "Test Plan",
			"description":     "Test Description",
			"tenor":           12,
			"status":          constant.InstallmentPlanStatusActive,
			"createdAt":       "0001-01-01T00:00:00Z",
			"updatedAt":       "0001-01-01T00:00:00Z",
			"deletedAt":       map[string]interface{}{"Time": "0001-01-01T00:00:00Z", "Valid": false},
			"planMetadata":    nil,
		},
	})

	testCases := []struct {
		name         string
		expectedCode int
		expectedBody string
		requestBody  func() []byte
		setupRequest func(req *http.Request) *http.Request
		setup        func(svc *mockSvc.IInstallmentPlanService)
	}{
		{
			name: "SUCCESS: Update installment plan",
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup: func(svc *mockSvc.IInstallmentPlanService) {
				svc.On(
					"Update",
					mock.Anything,
					mock.Anything,
				).Return(createdInstallmentPlan, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccess),
		},
		{
			name: "ERROR: Invalid ID",
			setupRequest: func(req *http.Request) *http.Request {
				return req
			},
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup:        func(svc *mockSvc.IInstallmentPlanService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"errors": "id is required"
			}`,
		},
		{
			name: "ERROR: Failed to decode request",
			requestBody: func() []byte {
				return []byte("invalid-json")
			},
			setup:        func(svc *mockSvc.IInstallmentPlanService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"errors": "invalid character 'i' looking for beginning of value"
			}`,
		},
		{
			name: "ERROR: Validation failed - invalid settlement type",
			requestBody: func() []byte {
				invalidPayload := *validPayload
				invalidPayload.SettlementType = "INVALID"
				body, _ := json.Marshal(invalidPayload)
				return body
			},
			setup:        func(svc *mockSvc.IInstallmentPlanService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"errors": {
					"SettlementType": "Key: 'UpdateInstallmentPlanRequest.SettlementType' Error:Field validation for 'SettlementType' failed on the 'oneof' tag"
				}
			}`,
		},
		{
			name: "ERROR: Service error",
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup: func(svc *mockSvc.IInstallmentPlanService) {
				svc.On(
					"Update",
					mock.Anything,
					mock.Anything,
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
			svc := mockSvc.NewIInstallmentPlanService(t)
			validator := validator.New()

			tc.setup(svc)

			ctrl := NewController(svc, validator)

			req := httptest.NewRequest(http.MethodPost, "/installment-plans/update", bytes.NewBuffer(tc.requestBody()))
			req = req.WithContext(ctx)
			if tc.setupRequest == nil {
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("id", uuid.NewString())
				req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, chiCtx))
			} else {
				req = tc.setupRequest(req)
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.Update)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())
		})
	}
}
