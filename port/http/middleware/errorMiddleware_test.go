package middleware_test

import (
	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetV2ErrorCodeMiddleware(pt *testing.T) {
	router := chi.NewRouter()
	MountHandlers(router, middleware.SetV2ErrorCodeMiddleware())

	tests := []struct {
		name           string
		reqSetting     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "SUCCESS",
			reqSetting: func(req *http.Request) {
				req.Header.Set("X-CLIENT-KEY", "merchant-id")
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

func TestSetErrorSourceMiddleware(pt *testing.T) {
	router := chi.NewRouter()
	MountHandlers(router, middleware.SetErrorSourceMiddleware())

	tests := []struct {
		name           string
		reqSetting     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "SUCCESS: middleware passes through and does not alter response",
			reqSetting: func(req *http.Request) {
				req.Header.Set("X-CLIENT-KEY", "merchant-id")
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
