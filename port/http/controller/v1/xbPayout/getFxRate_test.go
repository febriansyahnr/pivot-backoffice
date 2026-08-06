package xbPayoutController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/xbPayout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetFxRate(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, &userModel.UserTokenClaims{
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
			name: "ERROR: Invalid user info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"41", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"user not found"}`,
		},
		{
			name: "ERROR: Empty source currency",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"40", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"invalid source currency"}`,
		},
		{
			name: "ERROR: Empty destination currency",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			sourceCurrency:   "IDR",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"40", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"invalid destination currency"}`,
		},
		{
			name: "ERROR: GetFxRate service error",
			mockSetup: func() {
				xbPayoutSvc.On("GetFxRate",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetFxRateRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:          validRequestID,
			sourceCurrency:      "IDR",
			destinationCurrency: "USD",
			expectedStatus:      http.StatusInternalServerError,
			expectedRespBody:    `{"code":"99", "data":null, "error":{"details":[], "traceId":"", "type":"UNKNOWN"}, "message":"some error"}`,
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
			expectedRespBody:    `{"code":"00", "data":{"destinationFxRate":"0", "expiryAt":"0001-01-01T00:00:00Z", "fxRate":"0"}, "message":"OK"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/fx-rate?sourceCurrency=%s&destinationCurrency=%s", tc.sourceCurrency, tc.destinationCurrency), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/fx-rate", ctrl.GetFxRate)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}
