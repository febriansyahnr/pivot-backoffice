package middleware_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"errors"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckSubMerchantMiddleware(t *testing.T) {

	merchantInfo := &merchant.MerchantAuthTokenClaims{
		MerchantId: "aec6636d-7a02-4d93-a4c5-006b9c235068", // NOSONAR
	}

	subMerchant := &merchant.Merchant{
		ParentID: sql.NullString{
			String: merchantInfo.MerchantId,
			Valid:  true,
		},
	}

	tests := []struct {
		name           string
		reqSetting     func(req *http.Request)
		merchantInfo   *merchant.MerchantAuthTokenClaims
		mockSetup      func(merchantSvc *serviceMocks.IMerchantService, productSvc *serviceMocks.IProductService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: Empty Merchant Info",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			merchantInfo:   nil,
			mockSetup:      func(merchantSvc *serviceMocks.IMerchantService, productSvc *serviceMocks.IProductService) {},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"invalid token","error":{"type":"API_ERROR","message":"invalid token","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid Submerchant ID",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, "abced")
			},
			merchantInfo:   merchantInfo,
			mockSetup:      func(merchantSvc *serviceMocks.IMerchantService, productSvc *serviceMocks.IProductService) {},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"incorrect merchant id","error":{"type":"API_ERROR","message":"incorrect merchant id","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Find Merchant By ID",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			merchantInfo: merchantInfo,
			mockSetup: func(merchantSvc *serviceMocks.IMerchantService, productSvc *serviceMocks.IProductService) {
				merchantSvc.
					On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string")).
					Return(nil, pkgErrors.New(response.HttpErrRequest, errors.New("failed validation")))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"failed validation","error":{"type":"API_ERROR","message":"failed validation","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Empty Merchant",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			merchantInfo: merchantInfo,
			mockSetup: func(merchantSvc *serviceMocks.IMerchantService, productSvc *serviceMocks.IProductService) {
				merchantSvc.On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string")).Return(nil, nil)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"incorrect merchant id","error":{"type":"API_ERROR","message":"incorrect merchant id","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Mismatch ParentId",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			merchantInfo: merchantInfo,
			mockSetup: func(merchantSvc *serviceMocks.IMerchantService, productSvc *serviceMocks.IProductService) {
				merchantSvc.On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string")).Return(&merchant.Merchant{
					ParentID: sql.NullString{
						String: uuid.NewString(),
						Valid:  true,
					},
				}, nil)
			},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"43","message":"forbidden access","error":{"type":"API_ERROR","message":"forbidden access","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Product not available",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			merchantInfo: merchantInfo,
			mockSetup: func(merchantSvc *serviceMocks.IMerchantService, productSvc *serviceMocks.IProductService) {
				merchantSvc.On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string")).Return(subMerchant, nil)

				productSvc.On(
					"ValidateMerchantProductAvailability",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(pkgErrors.New(response.HttpErrForbidden, errors.New("errors")))
			},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"43","message":"errors","error":{"type":"API_ERROR","message":"errors","recommendation":""},"data":null}`,
		},
		{
			name:           "SUCCESS: No SubMerchant",
			reqSetting:     func(req *http.Request) {},
			merchantInfo:   merchantInfo,
			mockSetup:      func(merchantSvc *serviceMocks.IMerchantService, productSvc *serviceMocks.IProductService) {},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`,
		},
		{
			name: "SUCCESS: Check SubMerchant",
			reqSetting: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.NewString())
			},
			merchantInfo: merchantInfo,
			mockSetup: func(merchantSvc *serviceMocks.IMerchantService, productSvc *serviceMocks.IProductService) {
				merchantSvc.On("FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string")).Return(subMerchant, nil)
				productSvc.On("ValidateMerchantProductAvailability", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merchantSvc := serviceMocks.NewIMerchantService(t)
			productSvc := serviceMocks.NewIProductService(t)
			router := chi.NewRouter()
			MountHandlers(router, middleware.CheckSubMerchantMiddleware(merchantSvc, productSvc))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			ctx := context.WithValue(req.Context(), constant.CtxMerchantIDKey, merchantInfo.MerchantId)
			req = req.WithContext(context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantInfo))

			test.reqSetting(req)
			test.mockSetup(merchantSvc, productSvc)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, test.wantRespBody, rec.Body.String(), "Response body not matched. Want: %s, got: %s", test.wantRespBody, rec.Body.String())

		})
	}

}
