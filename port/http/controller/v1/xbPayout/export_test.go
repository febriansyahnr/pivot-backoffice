package xbPayoutController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestExportToExcel(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validMerchantID := "12345"
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, &userModel.UserTokenClaims{
			MerchantId: validMerchantID,
		}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(c.HeaderTimeZoneKey, "Asia/Jakarta")
	}

	validResponse := &xbModel.ExportXbPayoutResponse{
		Url: "https://storage.googleapis.com/bucket/xb-payout-history.xlsx?signed=true",
	}

	tests := []struct {
		name             string
		requestBody      string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name:        "ERROR: Invalid user info",
			requestBody: `{}`,
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"41", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"user not found"}`,
		},
		{
			name:        "ERROR: Invalid request body",
			requestBody: `invalid json`,
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"40", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"invalid character 'i' looking for beginning of value"}`,
		},
		{
			name:        "ERROR: ExportToExcel service error",
			requestBody: `{"startAt":"2025-09-08T00:00:00+07:00","endAt":"2025-10-09T23:59:59+07:00"}`,
			mockSetup: func() {
				xbPayoutSvc.On("ExportToExcel",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.ExportXbPayoutRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"99", "data":null, "error":{"details":[], "traceId":"", "type":"UNKNOWN"}, "message":"some error"}`,
		},
		{
			name:        "SUCCESS: Export with all filters",
			requestBody: `{"startAt":"2025-09-08T00:00:00+07:00","endAt":"2025-10-09T23:59:59+07:00","status":"SUCCESS","uuid":"test-uuid"}`,
			mockSetup: func() {
				xbPayoutSvc.On("ExportToExcel",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.ExportXbPayoutRequest"),
				).Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"url":"https://storage.googleapis.com/bucket/xb-payout-history.xlsx?signed=true"}, "message":"OK"}`,
		},
		{
			name:        "SUCCESS: Export with minimal filters",
			requestBody: `{"startAt":"2025-09-08T00:00:00+07:00","endAt":"2025-10-09T23:59:59+07:00"}`,
			mockSetup: func() {
				xbPayoutSvc.On("ExportToExcel",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.ExportXbPayoutRequest"),
				).Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"url":"https://storage.googleapis.com/bucket/xb-payout-history.xlsx?signed=true"}, "message":"OK"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/xb/payout/export", strings.NewReader(tc.requestBody))

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Post("/api/v1/xb/payout/export", ctrl.ExportToExcel)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}
