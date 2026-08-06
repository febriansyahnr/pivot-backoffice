package crmPaymentMethodController

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestActivatePaymentMethodMerchant(t *testing.T) {
	svc := serviceMocks.NewIPaymentMethodService(t)

	router := chi.NewRouter()
	router.Patch("/merchants/{id}/payment-methods/{paymentMethodId}/activate", New(svc).ActivatePaymentMethodMerchant)

	validMerchantID := uuid.NewString()
	validPaymentMethodID := uuid.NewString()

	tests := []struct {
		name            string
		merchantID      string
		paymentMethodID string
		modifierMock    func()
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name: "ERROR: Invalid merchantID format",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"id is required"}`,
		},
		{
			name:       "ERROR: Invalid paymentMethodID format",
			merchantID: validMerchantID,
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"paymentMethodId is required"}`,
		},
		{
			name:            "ERROR: Activate service error",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			modifierMock: func() {
				svc.On(
					"Activate",
					constant.ValueCtxMockType(),
					constant.PtrPaymentMethodWithPivot(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:            "SUCCESS",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			modifierMock: func() {
				svc.On(
					"Activate",
					constant.ValueCtxMockType(),
					constant.PtrPaymentMethodWithPivot(),
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"merchantId": "` + validMerchantID + `", "paymentMethodId": "` + validPaymentMethodID + `"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/merchants/%s/payment-methods/%s/activate", test.merchantID, test.paymentMethodID), nil)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestActivateAllPaymentMethod(t *testing.T) {
	svc := serviceMocks.NewIMerchantService(t)

	validMerchantID := uuid.NewString()

	tests := []struct {
		name           string
		merchantID     string
		environment    string
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: when the environment is production",
			environment: constant.EnvironmentProduction,
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"43","errors":"forbidden access"}`,
		},
		{
			name:       "ERROR: Invalid merchantID format",
			merchantID: "invalid",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"id is required"}`,
		},
		{
			name:       "ERROR: FindMerchantByID service error",
			merchantID: validMerchantID,
			modifierMock: func() {
				svc.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:       "ERROR: EnableAllPaymentMethod service error",
			merchantID: validMerchantID,
			modifierMock: func() {
				svc.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Once().Return(&merchant.Merchant{}, nil)

				svc.On(
					"EnableAllPaymentMethod",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:       "SUCCESS",
			merchantID: validMerchantID,
			modifierMock: func() {
				svc.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Once().Return(&merchant.Merchant{}, nil)

				svc.On(
					"EnableAllPaymentMethod",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"merchantId": "` + validMerchantID + `", "paymentMethodId": "all"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			cfg := &config.Config{
				Environment: constant.EnvironmentStaging,
			}

			if test.environment != "" {
				cfg.Environment = test.environment
			}

			handler := New(nil, WithMerchantService(svc), WithConfig(cfg))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/merchants/%s/payment-methods/activate", test.merchantID), nil)

			router := chi.NewRouter()
			router.Patch("/merchants/{id}/payment-methods/activate", handler.ActivateAllPaymentMethod)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
