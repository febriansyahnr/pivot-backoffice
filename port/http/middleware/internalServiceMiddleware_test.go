package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternalServiceMiddleware(pt *testing.T) {
	router := chi.NewRouter()
	secret := &config.Secret{
		InternalApiKeySecret: config.InternalApiKeySecret{
			Salt:       "l4,Ef7zv/N>ROHd79__rN7mW&n+Kibm?B`8026d~S9)fIduHLm(jIsec8A1K!j]3",
			HashResult: "6beaa3de45585022c553c653589ec20cca4794bd7e8961004cd31da7f4619460",
		},
	}
	MountHandlers(router, middleware.InternalServiceMiddleware(secret))

	tests := []struct {
		name           string
		reqSetting     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Token not sent",
			reqSetting:     func(*http.Request) {},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code": "credentials_invalid","error": {"details": [{"field": "", "message": "Unauthorized"}],"traceId": "", "type": "API_ERROR" },"message": "Unauthorized"}`,
		},
		{
			name: "ERROR:Invalid API key",
			reqSetting: func(req *http.Request) {
				req.Header.Set("INTERNAL-API-KEY", "invalid-api-key")
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code": "credentials_invalid","error": {"details": [{"field": "", "message": "Unauthorized"}],"traceId": "", "type": "API_ERROR" },"message": "Unauthorized"}`,
		},
		{
			name: "SUCCESS",
			reqSetting: func(req *http.Request) {
				req.Header.Set("INTERNAL-API-KEY", "042bbda8-464b-4a1a-8eb3-df2db7f93504")
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			test.reqSetting(req)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
