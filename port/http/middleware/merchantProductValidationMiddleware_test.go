package middleware_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	chi "github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMerchantUserProductValidationMiddleware(t *testing.T) {
	userClaims := &userModel.UserTokenClaims{}
	testCases := []struct {
		name         string
		setup        func(productSvc *serviceMocks.IProductService)
		setupRequest func(req *http.Request) *http.Request
		expectedCode int
	}{
		{
			name: "SUCCESS",
			setupRequest: func(req *http.Request) *http.Request {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims))
				return req
			},
			setup: func(productSvc *serviceMocks.IProductService) {
				productSvc.On(
					"ValidateMerchantProductAvailability",
					mock.Anything, mock.Anything,
				).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "SUCCESS: Validation failed",
			setupRequest: func(req *http.Request) *http.Request {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims))
				return req
			},
			setup: func(productSvc *serviceMocks.IProductService) {
				errMsg := fmt.Sprintf(constant.MerchantIsNotAllowedToUseProductMsgFormat, "PRODUCT")
				productSvc.On(
					"ValidateMerchantProductAvailability",
					mock.Anything, mock.Anything,
				).Return(
					pkgErrors.New(response.HttpErrForbidden, errors.New(errMsg)),
				)
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "ERROR: No claims",
			setupRequest: func(req *http.Request) *http.Request {
				return req
			},
			setup: func(productSvc *serviceMocks.IProductService) {
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "ERROR: Error service",
			setupRequest: func(req *http.Request) *http.Request {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims))
				return req
			},
			setup: func(productSvc *serviceMocks.IProductService) {
				productSvc.On(
					"ValidateMerchantProductAvailability",
					mock.Anything, mock.Anything,
				).Return(errors.New("errors"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			productSvc := serviceMocks.NewIProductService(t)
			tc.setup(productSvc)

			router := chi.NewRouter()
			MountHandlers(router, middleware.MerchantUserProductValidationMiddleware(productSvc, "test"))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			req = tc.setupRequest(req)

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.expectedCode, rec.Result().StatusCode)
		})
	}
}

func TestMerchantProductValidationMiddleware(t *testing.T) {
	merchantClaims := &merchant.MerchantAuthTokenClaims{}
	testCases := []struct {
		name         string
		setup        func(productSvc *serviceMocks.IProductService)
		setupRequest func(req *http.Request) *http.Request
		expectedCode int
	}{
		{
			name: "SUCCESS",
			setupRequest: func(req *http.Request) *http.Request {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, merchantClaims))
				return req
			},
			setup: func(productSvc *serviceMocks.IProductService) {
				productSvc.On(
					"ValidateMerchantProductAvailability",
					mock.Anything, mock.Anything,
				).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "SUCCESS: Validation failed",
			setupRequest: func(req *http.Request) *http.Request {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, merchantClaims))
				return req
			},
			setup: func(productSvc *serviceMocks.IProductService) {
				errMsg := fmt.Sprintf(constant.MerchantIsNotAllowedToUseProductMsgFormat, "PRODUCT")
				productSvc.On(
					"ValidateMerchantProductAvailability",
					mock.Anything, mock.Anything,
				).Return(
					pkgErrors.New(response.HttpErrForbidden, errors.New(errMsg)),
				)
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "ERROR: No claims",
			setupRequest: func(req *http.Request) *http.Request {
				return req
			},
			setup: func(productSvc *serviceMocks.IProductService) {
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "ERROR: Error service",
			setupRequest: func(req *http.Request) *http.Request {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, merchantClaims))
				return req
			},
			setup: func(productSvc *serviceMocks.IProductService) {
				productSvc.On(
					"ValidateMerchantProductAvailability",
					mock.Anything, mock.Anything,
				).Return(errors.New("errors"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			productSvc := serviceMocks.NewIProductService(t)
			tc.setup(productSvc)

			router := chi.NewRouter()
			MountHandlers(router, middleware.MerchantProductValidationMiddleware(productSvc, "test"))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			req = tc.setupRequest(req)

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.expectedCode, rec.Result().StatusCode)
		})
	}
}
