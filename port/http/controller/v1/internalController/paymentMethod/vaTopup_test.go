package internalPaymentMethodController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTopUpVAPaymentMethod(t *testing.T) {
	merchantTopupResult := &model.MerchantTopUp{
		ID:              uuid.Max.String(),
		MerchantID:      uuid.Max.String(),
		PaymentMethodID: uuid.Max.String(),
		ReferenceNumber: "reference number",
	}

	testCases := []struct {
		name             string
		merchantId       string
		expectedStatus   int
		expectedResponse string
		claimMerchant    bool
		mockSetup        func(topUpSvc *mockSvc.IMerchantTopUpService)
		setHeaders       func(req *http.Request)
		setParams        func(chiCtx *chi.Context)
	}{
		{
			name:             "SUCCESS: Response 200",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"uuid":"ffffffff-ffff-ffff-ffff-ffffffffffff","merchantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","paymentMethodId":"ffffffff-ffff-ffff-ffff-ffffffffffff","referenceNumber":"reference number","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
			claimMerchant:    true,
			mockSetup: func(topUpSvc *mockSvc.IMerchantTopUpService) {
				topUpSvc.On(
					"FindOrCreate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(merchantTopupResult, nil)
			},
			setParams: func(chiCtx *chi.Context) { chiCtx.URLParams.Add("paymentMethodId", uuid.NewString()) },
		},
		{
			name:             "SUCCESS: Response 200 with Submerchant ID",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"uuid":"ffffffff-ffff-ffff-ffff-ffffffffffff","merchantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","paymentMethodId":"ffffffff-ffff-ffff-ffff-ffffffffffff","referenceNumber":"reference number","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
			claimMerchant:    true,
			mockSetup: func(topUpSvc *mockSvc.IMerchantTopUpService) {
				topUpSvc.On(
					"FindOrCreate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(merchantTopupResult, nil)
			},
			setHeaders: func(req *http.Request) { req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString()) },
			setParams:  func(chiCtx *chi.Context) { chiCtx.URLParams.Add("paymentMethodId", uuid.NewString()) },
		},
		{
			name:             "ERROR: Invalid merchant claims",
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: `{"code":"41","message":"merchant not found","error":{"type":"API_ERROR","message":"merchant not found","recommendation":""},"data":null}`,
			mockSetup:        func(_ *mockSvc.IMerchantTopUpService) { /* Empty Body Function */ },
			setHeaders:       func(_ *http.Request) { /* Empty Body Function */ },
		},
		{
			name:             "ERROR: Empty payment method id",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":"40","message":"invalid payment method id","error":{"type":"API_ERROR","message":"invalid payment method id","recommendation":""},"data":null}`,
			mockSetup:        func(_ *mockSvc.IMerchantTopUpService) { /* Empty Body Function */ },
			setHeaders:       func(req *http.Request) { /* Empty Body Function */ },
			claimMerchant:    true,
		},
		{
			name:             "ERROR: Empty payment method id (new response)",
			merchantId:       "a0e4d057-3712-4e05-ba1e-e632bc28a544", // NOSONAR
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":"field_required","message":"Mandatory field is missing","error":{"type":"API_ERROR","details":[{"field":"paymentMethodId","message":"Make sure paymentMethodId value is fulfilled"}],"traceId":""}}`,
			mockSetup:        func(_ *mockSvc.IMerchantTopUpService) { /* Empty Body Function */ },
			setHeaders:       func(req *http.Request) { /* Empty Body Function */ },
			claimMerchant:    true,
		},
		{
			name:             "ERROR: Error Find or Create",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"some error","error":{"type":"UNKNOWN","message":"some error","recommendation":""},"data":null}`,
			claimMerchant:    true,
			mockSetup: func(topUpSvc *mockSvc.IMerchantTopUpService) {
				topUpSvc.On(
					"FindOrCreate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)

			},
			setParams: func(chiCtx *chi.Context) { chiCtx.URLParams.Add("paymentMethodId", uuid.NewString()) },
		},
		{
			name:             "ERROR: Partner downstream (snap-core bad gateway) propagates as 502",
			merchantId:       "a0e4d057-3712-4e05-ba1e-e632bc28a544", // NOSONAR
			expectedStatus:   http.StatusBadGateway,
			expectedResponse: `{"code":"bad_gateway","message":"An internal error was encountered. Please Try again later","error":{"type":"GATEWAY_ERROR","details":[{"field":"","message":"assert.AnError general error for testing"}],"traceId":""}}`,
			claimMerchant:    true,
			mockSetup: func(topUpSvc *mockSvc.IMerchantTopUpService) {
				topUpSvc.On(
					"FindOrCreate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil, pkgErrs.New(response.HttpErrBadGateway, assert.AnError))

			},
			setParams: func(chiCtx *chi.Context) { chiCtx.URLParams.Add("paymentMethodId", uuid.NewString()) },
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			topUpSvc := mockSvc.NewIMerchantTopUpService(t)
			tc.mockSetup(topUpSvc)

			ctrl := New(topUpSvc, nil)

			baseUrl := "/open-api/v1/payment-methods/virtual-accounts/" + uuid.NewString() + "/top-up"
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			chiRouterCtx := chi.NewRouteContext()

			if tc.setParams != nil {
				tc.setParams(chiRouterCtx)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			}

			if tc.merchantId == "" {
				tc.merchantId = "aec6636d-7a02-4d93-a4c5-006b9c235068" // NOSONAR
			}

			ctx := context.WithValue(req.Context(), constant.CtxMerchantIDKey, tc.merchantId)
			if tc.claimMerchant {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, &merchant.MerchantAuthTokenClaims{MerchantId: tc.merchantId})
			}
			req = req.WithContext(ctx)

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.TopUpVAPaymentMethod)
			handler.ServeHTTP(httpRecorder, req)

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			if !assert.JSONEq(t, tc.expectedResponse, httpRecorder.Body.String()) {
				t.Log("Result:", httpRecorder.Body.String())
			}
		})
	}
}
