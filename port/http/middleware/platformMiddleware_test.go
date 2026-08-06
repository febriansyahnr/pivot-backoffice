package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestPlatformDepositSetting(t *testing.T) {
	permissionService := serviceMocks.NewIPermissionService(t)

	router := chi.NewRouter()
	router.Use(PlatformPermissionOption(
		PermittedPermissions(permissionService, c.PermissionSlugDepositSettingView), permissionService, c.PermissionSlugPlatformView),
	)
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message":"OK"}`))
	})

	userClaims := &user.UserTokenClaims{RoleID: "6695fa73-75c4-4043-8395-32858798be5a"}

	tests := []struct {
		name           string
		setupRequest   func(r *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR:Request not permitted",
			setupRequest: func(r *http.Request) {
				r.Header.Set(c.HeaderXSubMerchantID, "12345")

				permissionService.On(
					"GetCachedPermissionsByRoleId", c.ValueCtxMockType(), "6695fa73-75c4-4043-8395-32858798be5a",
				).Once().Return([]string{c.PermissionSlugPaymentInsightView}, nil)
			},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   `{"code":"43","message":"forbidden access","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "SUCCESS:Platform",
			setupRequest: func(r *http.Request) {
				r.Header.Set(c.HeaderXSubMerchantID, "12345")

				permissionService.On(
					"GetCachedPermissionsByRoleId", c.ValueCtxMockType(), "6695fa73-75c4-4043-8395-32858798be5a",
				).Once().Return([]string{c.PermissionSlugPlatformView}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`,
		},
		{
			name: "SUCCESS:Non-Platform",
			setupRequest: func(r *http.Request) {
				permissionService.On(
					"GetCachedPermissionsByRoleId", c.ValueCtxMockType(), "6695fa73-75c4-4043-8395-32858798be5a",
				).Once().Return([]string{c.PermissionSlugDepositSettingView}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message":"OK"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			test.setupRequest(req)
			req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, userClaims))

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
