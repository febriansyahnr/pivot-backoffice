package internalPaymentMethodController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetVAPaymentMethods(t *testing.T) {
	description := "description"
	logo := "logo"
	bankName := "bankName"
	response := []*paymentModel.PaymentMethodWithPivot{
		{
			PaymentMethod: paymentModel.PaymentMethod{
				UUID:        "uuid",
				Type:        "type",
				Category:    "category",
				Name:        "name",
				Description: &description,
				Logo:        &logo,
				Acquirer:    "acquirer",
				BankName:    &bankName,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	merchantId := "aec6636d-7a02-4d93-a4c5-006b9c235068" // NOSONAR

	testCases := []struct {
		name             string
		merchantId       string
		expectedStatus   int
		expectedResponse string
		claimMerchant    bool
		mockSetup        func(paymentMethodSvc *mockSvc.IPaymentMethodService)
		setHeaders       func(req *http.Request)
	}{
		{
			name:             "SUCCESS: Response 200",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"Success","data":[{"uuid":"uuid","name":"name","description":"description","logo":"logo","acquirer":"acquirer","bankName":"bankName"}]}`,
			claimMerchant:    true,
			mockSetup: func(paymentMethodSvc *mockSvc.IPaymentMethodService) {
				paymentMethodSvc.On("GetPaymentMethodByMerchant", mock.Anything, mock.Anything).Return(response, nil)
			},
		},
		{
			name:             "SUCCESS: Response 200 with Submerchant ID",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"Success","data":[{"uuid":"uuid","name":"name","description":"description","logo":"logo","acquirer":"acquirer","bankName":"bankName"}]}`,
			claimMerchant:    true,
			mockSetup: func(paymentMethodSvc *mockSvc.IPaymentMethodService) {
				paymentMethodSvc.On("GetPaymentMethodByMerchant", mock.Anything, mock.Anything).Return(response, nil)
			},
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
		},
		{
			name:             "ERROR: Invalid merchant claims",
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: `{"code":"41","message":"merchant not found","error":{"type":"API_ERROR","message":"merchant not found","recommendation":""},"data":null}`,
			mockSetup:        func(_ *mockSvc.IPaymentMethodService) { /*Empty Body Function*/ },
			setHeaders:       func(_ *http.Request) { /*Empty Body Function*/ },
		},
		{
			name:             "ERROR: Error get payment methods",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"errors","error":{"type":"UNKNOWN","message":"errors","recommendation":""},"data":null}`,
			claimMerchant:    true,
			mockSetup: func(paymentMethodSvc *mockSvc.IPaymentMethodService) {
				paymentMethodSvc.On("GetPaymentMethodByMerchant", mock.Anything, mock.Anything).Return(nil, errors.New("errors")) // NOSONAR
			},
		},
		{
			name:             "ERROR: Error get payment methods (new error response)",
			merchantId:       "ecb67ae4-370c-4e5f-830d-a95d1b8b2943", // NOSONAR
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
			claimMerchant:    true,
			mockSetup: func(paymentMethodSvc *mockSvc.IPaymentMethodService) {
				paymentMethodSvc.On("GetPaymentMethodByMerchant", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentMethodSvc := mockSvc.NewIPaymentMethodService(t)
			tc.mockSetup(paymentMethodSvc)

			ctrl := New(nil, paymentMethodSvc)

			baseUrl := "/open-api/v1/payment-methods/virtual-accounts"
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			chiRouterCtx := chi.NewRouteContext()

			if tc.merchantId == "" {
				tc.merchantId = merchantId
			}
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx)
			ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, tc.merchantId)

			if tc.claimMerchant {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{
					MerchantId: uuid.NewString(),
				})
			}
			req = req.WithContext(ctx)

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.GetVAPaymentMethods)
			handler.ServeHTTP(httpRecorder, req)

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			if !assert.JSONEq(t, tc.expectedResponse, httpRecorder.Body.String()) {
				t.Log("Result:", httpRecorder.Body.String())
			}
		})
	}
}
