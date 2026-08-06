package middleware_test

import (
	"context"
	"github.com/google/uuid"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestPermittedPermissions(t *testing.T) {
	permissionSvc := serviceMocks.NewIPermissionService(t)
	roleId := uuid.NewString()

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		wantStatusCode int
		wantRespBody   string
		setupMock      func()
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"user not found"}`,
			setupMock: func() {
				// empty setup mock
			},
		},
		{
			name:           "ERROR:Access forbidden",
			userClaims:     &user.UserTokenClaims{RoleID: roleId},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"43","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"forbidden access"}`,
			setupMock: func() {
				permissionSvc.On(
					"GetCachedPermissionsByRoleId",
					constant.ValueCtxMockType(),
					roleId,
				).Once().Return([]string{}, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:           "SUCCESS",
			userClaims:     &user.UserTokenClaims{RoleID: roleId},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK-OK-OK"}`,
			setupMock: func() {
				permissionSvc.On(
					"GetCachedPermissionsByRoleId",
					constant.ValueCtxMockType(),
					roleId,
				).Once().Return([]string{constant.PermissionSlugDeveloperSettingView}, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/settings", nil)

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}

			if test.setupMock != nil {
				test.setupMock()
			}

			router := chi.NewRouter()
			router.Use(middleware.PermittedPermissions(permissionSvc, constant.PermissionSlugDeveloperSettingView))
			router.Get("/settings", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set(constant.HeaderContentType, constant.MIMEApplicationJSON)
				w.WriteHeader(http.StatusOK)

				_, _ = w.Write([]byte(`{"message":"OK-OK-OK"}`))
			})

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
