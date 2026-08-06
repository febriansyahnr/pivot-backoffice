package crmPaymentMethodController

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"

	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestGetByMerchant(t *testing.T) {
	svc := serviceMocks.NewIPaymentMethodService(t)

	validMerchantID := uuid.NewString()
	validResponse := []*paymentModel.PaymentMethodWithPivot{
		{
			MerchantID: validMerchantID,
			IsActive:   true,
		},
	}

	validResponseInJson, err := json.Marshal(validResponse)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	tests := []struct {
		name           string
		merchantID     string
		queryParams    map[string]string
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
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
			name:       "ERROR: GetPaymentMethodByMerchant service error",
			merchantID: validMerchantID,
			modifierMock: func() {
				svc.On(
					"GetPaymentMethodByMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:       "SUCCESS: without query parameters",
			merchantID: validMerchantID,
			modifierMock: func() {
				svc.On(
					"GetPaymentMethodByMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(validResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseInJson) + `}`,
		},
		{
			name:       "SUCCESS: with query parameters",
			merchantID: validMerchantID,
			queryParams: map[string]string{
				"category": "PAYMENT",
				"type":     "QRIS",
				"acquirer": "BNI",
				"subtype":  "subtype",
			},
			modifierMock: func() {
				svc.On(
					"GetPaymentMethodByMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(validResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseInJson) + `}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/merchants/%s/payment-methods", test.merchantID)

			// Add query parameters if provided
			if len(test.queryParams) > 0 {
				url += "?"
				for key, value := range test.queryParams {
					url += fmt.Sprintf("%s=%s&", key, value)
				}
				url = url[:len(url)-1] // Remove trailing &
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)

			router := chi.NewRouter()
			router.Get("/merchants/{id}/payment-methods", New(svc).GetByMerchant)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestGetStaticVAByMerchant(t *testing.T) {
	svc := serviceMocks.NewIPaymentMethodService(t)

	validMerchantID := uuid.NewString()
	validResponse := []*paymentModel.PaymentMethodWithPivot{
		{
			MerchantID: validMerchantID,
			IsActive:   true,
		},
	}

	validResponseInJson, err := json.Marshal(validResponse)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	tests := []struct {
		name           string
		merchantID     string
		queryParams    map[string]string
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
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
			name:       "ERROR: GetStaticVAPaymentMethodByMerchant service error",
			merchantID: validMerchantID,
			modifierMock: func() {
				svc.On(
					"GetStaticVAPaymentMethodByMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:       "SUCCESS: without query parameters",
			merchantID: validMerchantID,
			modifierMock: func() {
				svc.On(
					"GetStaticVAPaymentMethodByMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(validResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseInJson) + `}`,
		},
		{
			name:       "SUCCESS: with acquirer query parameter",
			merchantID: validMerchantID,
			queryParams: map[string]string{
				"acquirer": "BCA",
			},
			modifierMock: func() {
				svc.On(
					"GetStaticVAPaymentMethodByMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(validResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseInJson) + `}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/merchants/%s/payment-methods/static-va", test.merchantID)

			// Add query parameters if provided
			if len(test.queryParams) > 0 {
				url += "?"
				for key, value := range test.queryParams {
					url += fmt.Sprintf("%s=%s&", key, value)
				}
				url = url[:len(url)-1] // Remove trailing &
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)

			router := chi.NewRouter()
			router.Get("/merchants/{id}/payment-methods/static-va", New(svc).GetStaticVAByMerchant)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestGetStaticQRByMerchant(t *testing.T) {
	svc := serviceMocks.NewIPaymentMethodService(t)

	validMerchantID := uuid.NewString()
	validResponse := &paymentModel.PaymentMethodWithPivot{
		MerchantID: validMerchantID,
		IsActive:   true,
		QRPayments: []paymentModel.StaticQRPaymentItem{
			{
				MerchantID:               validMerchantID,
				PaymentSessionID:         "payment-123",
				PaymentClientReferenceID: "ref-123",
				StoreID:                  "store-123",
				IsDerived:                false,
			},
		},
	}

	validResponseInJson, err := json.Marshal(validResponse)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	tests := []struct {
		name           string
		merchantID     string
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
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
			name:       "ERROR: GetStaticQRPaymentMethodByMerchant service error",
			merchantID: validMerchantID,
			modifierMock: func() {
				svc.On(
					"GetStaticQRPaymentMethodByMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:       "SUCCESS: Get static QR payment method",
			merchantID: validMerchantID,
			modifierMock: func() {
				svc.On(
					"GetStaticQRPaymentMethodByMerchant",
					constant.ValueCtxMockType(),
					constant.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(validResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseInJson) + `}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/merchants/%s/payment-methods/static-qris", test.merchantID)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			router := chi.NewRouter()
			router.Get("/merchants/{id}/payment-methods/static-qris", New(svc).GetStaticQRByMerchant)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
