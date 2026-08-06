package installmentplan

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	merchantID := uuid.New().String()
	midID := uuid.New().String()

	validPayload := &installmentPlanModel.CreateInstallmentPlanRequest{
		MerchantID:     merchantID,
		Acquirer:       "HARSYA",
		SettlementType: "AGGREGATOR",
		PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
		Title:          "Test Plan",
		Description:    "Test Description",
		Tenor:          12,
		CardDetail: &installmentPlanModel.CardInstallmentPlanRequest{
			MidID:         midID,
			AllowedBins:   []string{"123456", "654321"},
			Interest:      2.5,
			MinimumAmount: 1000,
			MaximumAmount: 100000,
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
		setup        func(svc *mockSvc.IInstallmentPlanService)
	}{
		{
			name: "SUCCESS: Create installment plan",
			requestBody: func() []byte {
				body, _ := json.Marshal(validPayload)
				return body
			},
			setup: func(svc *mockSvc.IInstallmentPlanService) {
				svc.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*installmentPlanModel.CreateInstallmentPlanRequest"),
				).Return(createdInstallmentPlan, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: string(expectedSuccess),
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
			name: "ERROR: Validation failed - missing title",
			requestBody: func() []byte {
				invalidPayload := *validPayload
				invalidPayload.Title = ""
				body, _ := json.Marshal(invalidPayload)
				return body
			},
			setup:        func(svc *mockSvc.IInstallmentPlanService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"errors": {
					"Title": "Key: 'CreateInstallmentPlanRequest.Title' Error:Field validation for 'Title' failed on the 'required' tag"
				}
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
					"SettlementType": "Key: 'CreateInstallmentPlanRequest.SettlementType' Error:Field validation for 'SettlementType' failed on the 'oneof' tag"
				}
			}`,
		},
		{
			name: "ERROR: Validation failed - tenor too small",
			requestBody: func() []byte {
				invalidPayload := *validPayload
				invalidPayload.Tenor = 0
				body, _ := json.Marshal(invalidPayload)
				return body
			},
			setup:        func(svc *mockSvc.IInstallmentPlanService) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{
				"code": "40",
				"errors": {
					"Tenor": "Key: 'CreateInstallmentPlanRequest.Tenor' Error:Field validation for 'Tenor' failed on the 'required' tag"
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
					"Create",
					mock.Anything,
					mock.AnythingOfType("*installmentPlanModel.CreateInstallmentPlanRequest"),
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

			req := httptest.NewRequest(http.MethodPost, "/installment-plans/create", bytes.NewBuffer(tc.requestBody()))
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(ctrl.Create)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())
		})
	}
}
