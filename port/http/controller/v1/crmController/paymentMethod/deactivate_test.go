package crmPaymentMethodController

import (
	"fmt"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestDeactivatePaymentMethodMerchant(t *testing.T) {
	svc := serviceMocks.NewIPaymentMethodService(t)

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
			name:            "ERROR: Invalid merchantID format",
			merchantID:      "invalid",
			paymentMethodID: "",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"id is required"}`,
		},
		{
			name:            "ERROR: Invalid paymentMethodID format",
			merchantID:      validMerchantID,
			paymentMethodID: "invalid",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"paymentMethodId is required"}`,
		},
		{
			name:            "ERROR: FindPaymentMethodByIdAndMerchant service error",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			modifierMock: func() {
				svc.On(
					"FindPaymentMethodByIdAndMerchant",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:            "ERROR: Payment status already inactive",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			modifierMock: func() {
				svc.On(
					"FindPaymentMethodByIdAndMerchant",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Once().Return(&paymentModel.PaymentMethodWithPivot{
					IsActive: false,
				}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"payment method is already inactive"}`,
		},
		{
			name:            "ERROR: Deactivate service error",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			modifierMock: func() {
				svc.On(
					"FindPaymentMethodByIdAndMerchant",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&paymentModel.PaymentMethodWithPivot{
					IsActive: true,
				}, nil)

				svc.On(
					"Deactivate",
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
					"Deactivate",
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
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/merchants/%s/payment-methods/%s/deactivate", test.merchantID, test.paymentMethodID), nil)

			router := chi.NewRouter()
			router.Patch("/merchants/{id}/payment-methods/{paymentMethodId}/deactivate", New(svc).DeactivatePaymentMethodMerchant)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
