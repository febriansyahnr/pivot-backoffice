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

func TestEnvironmentCheck(t *testing.T) {
	tests := []struct {
		name           string
		currentEnv     string
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Env local",
			currentEnv:     constant.EnvironmentLocal,
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"43", "errors": "forbidden access"}`,
		},
		{
			name:           "ERROR:Env production",
			currentEnv:     constant.EnvironmentProduction,
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"43", "errors": "forbidden access"}`,
		},
		{
			name:           "SUCCESS:Env staging",
			currentEnv:     constant.EnvironmentStaging,
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message": "OK"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Use(middleware.EnvironmentCheck(test.currentEnv, constant.EnvironmentStaging))
			router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				_, _ = w.Write([]byte(`{"message": "OK"}`))
				w.WriteHeader(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
