package middleware_test

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redismock/v9"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIdempotencyMiddleware(pt *testing.T) {
	addReq := func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer dummy-token")
		req.Header.Set("X-Idempotent-Key", "12345")
	}
	tests := []struct {
		name           string
		reqSetting     func(req *http.Request)
		mockSetup      func(r redismock.ClientMock)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:       "SUCCESS",
			reqSetting: addReq,
			mockSetup: func(r redismock.ClientMock) {
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message": "OK"}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			db, clientMock := redismock.NewClientMock()
			redisMock := redisExt.WrapRedisClient(db, nil)

			router := chi.NewRouter()

			test.reqSetting(req)
			test.mockSetup(clientMock)

			MountHandlers(router, middleware.IdempotencyMiddleware(redisMock, "-"))

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}

}
