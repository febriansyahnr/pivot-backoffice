package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyAuth(t *testing.T) {
	router := chi.NewRouter()

	authValues := []string{
		"AUTHOKR-09283",
		"HELLO-KWJEH#",
	}
	router.Use(middleware.KeyAuth(constant.HeaderXCRMKey, authValues...))
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_, _ = w.Write([]byte(`{"message": "OK"}`))
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		mockRequest    func(r *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR:Empty key auth",
			mockRequest: func(_ *http.Request) {
				// Empty mock request
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41", "errors": "key auth required"}`,
		},
		{
			name: "ERROR:Invalid key",
			mockRequest: func(r *http.Request) {
				r.Header.Set(constant.HeaderXCRMKey, "WRONG-API-KEY")
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41", "errors": "invalid key"}`,
		},
		{
			name: "SUCCESS:With API Key #1",
			mockRequest: func(r *http.Request) {
				r.Header.Set(constant.HeaderXCRMKey, authValues[0])
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`,
		},
		{
			name: "SUCCESS:With API Key #2",
			mockRequest: func(r *http.Request) {
				r.Header.Set(constant.HeaderXCRMKey, authValues[1])
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			test.mockRequest(req)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
