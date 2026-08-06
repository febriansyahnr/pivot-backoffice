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

func TestGetFxRate(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	svc := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	tests := []struct {
		name                string
		mockSetup           func()
		reqSetting          func(r *http.Request)
		sourceCurrency      string
		destinationCurrency string
		expectedStatus      int
		expectedRespBody    string
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
			name: "ERROR: Invalid source currency",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Invalid destination currency",
			mockSetup: func() {
				// empty modifier
			},
			sourceCurrency:   "IDR",
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: GetFxRate service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetFxRate",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetFxRateRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			sourceCurrency:      "IDR",
			destinationCurrency: "USD",
			reqSetting:          validRequestID,
			expectedStatus:      http.StatusInternalServerError,
			expectedRespBody:    `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("GetFxRate",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetFxRateRequest"),
				).Once().Return(&xbModel.GetFxRateResponse{}, nil)
			},
			reqSetting:          validRequestID,
			sourceCurrency:      "IDR",
			destinationCurrency: "USD",
			expectedStatus:      http.StatusOK,
			expectedRespBody:    `{"code":"00","data": {"destinationFxRate":"0", "expiryAt":"0001-01-01T00:00:00Z", "fxRate":"0"}, "message":"Success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/open-api/v1/xb/fx-rate?sourceCurrency=%s&destinationCurrency=%s", tt.sourceCurrency, tt.destinationCurrency), nil)

			if tt.reqSetting != nil {
				tt.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/open-api/v1/xb/fx-rate", svc.GetFxRate)

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tt.expectedRespBody, rec.Body.String())
		})
	}
}
