package bankController

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetList(t *testing.T) {
	testCases := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "SUCCESS: Get List without any filter",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{}
			bankController := New(cfg)
			baseUrl := "/api/v1/banks"

			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(bankController.List)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
		})
	}
}
