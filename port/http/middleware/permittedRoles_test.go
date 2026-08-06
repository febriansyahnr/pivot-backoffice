package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestPermittedRoles(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.PermittedRoles(constant.RoleDeveloper))
	router.Get("/settings", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(constant.HeaderContentType, constant.MIMEApplicationJSON)
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{"message":"OK-OK-OK"}`))
	})

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"user not found"}`,
		},
		{
			name:           "ERROR:Access forbidden",
			userClaims:     &user.UserTokenClaims{Role: constant.RoleMaker},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"43","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"forbidden access"}`,
		},
		{
			name:           "SUCCESS",
			userClaims:     &user.UserTokenClaims{Role: constant.RoleDeveloper},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK-OK-OK"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/settings", nil)

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
