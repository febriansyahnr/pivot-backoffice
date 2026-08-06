package internalXbController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/xb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetListMasterCountry(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterCountry",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterCountry",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterCountry)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListMasterCurrency(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterCurrency",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterCurrency",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterCurrency)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListMasterState(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterState",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterState",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterState)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListMasterCity(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterCity",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterCity",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterCity)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListMasterCurrencyMapping(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterCurrencyMapping",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterCurrencyMapping",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterCurrencyMapping)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListMasterIdentificationType(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterIdentificationType",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterIdentificationType",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterIdentificationType)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListMasterAccountType(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterAccountType",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterAccountType",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterAccountType)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListMasterPurpose(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterPurpose",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterPurpose",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterPurpose)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListMasterTransferMethod(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterTransferMethod",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterTransferMethod",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterTransferMethod)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListMasterSourceOfIncome(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &xbModel.PaginationResponse{}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		page             string
		perPage          string
		fetchAll         string
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid page input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid perPage input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid fetchAll input",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterSourceOfIncome",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			page:             "0",
			perPage:          "10",
			fetchAll:         "true",
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetListMasterSourceOfIncome",
					c.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/master/list?page=%s&perPage=%s&fetchAll=%s", tc.page, tc.perPage, tc.fetchAll), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/master/list", ctrl.GetListMasterSourceOfIncome)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}
