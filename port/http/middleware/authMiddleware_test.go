package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	jwt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	redis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"

	chi "github.com/go-chi/chi/v5"
	goJWT "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(pt *testing.T) {
	mockJwt := jwt.NewIJwt(pt)
	router := chi.NewRouter()
	MountHandlers(router, middleware.AuthMiddleware(mockJwt, &redis.IRedisExt{}))

	addReqToken := func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer access_token")
	}
	tests := []struct {
		name           string
		reqSetting     func(req *http.Request)
		mockSetup      func(jwtCore *jwt.IJwt)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:No Token",
			reqSetting:     func(*http.Request) {},
			mockSetup:      func(*jwt.IJwt) {},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid token"}`,
		},
		{
			name:       "ERROR:Token not recognized",
			reqSetting: addReqToken,
			mockSetup: func(jwtCore *jwt.IJwt) {
				jwtCore.On(
					"Verify", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, errors.New("token not recognized"))
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid token"}`,
		},
		{
			name:       "ERROR:Token has expired",
			reqSetting: addReqToken,
			mockSetup: func(jwtCore *jwt.IJwt) {
				jwtCore.On(
					"Verify", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(func() (*user.UserTokenClaims, error) {
					u := &user.UserTokenClaims{}
					u.RegisteredClaims.ExpiresAt = &goJWT.NumericDate{Time: time.Now().Add(-time.Minute).UTC()}
					return u, nil
				}())
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"token is expired"}`,
		},
		{
			name:       "ERROR:Get token logged in device",
			reqSetting: addReqToken,
			mockSetup: func(jwtCore *jwt.IJwt) {
				jwtCore.On(
					"Verify", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(func() (*user.UserTokenClaims, error) {
					u := &user.UserTokenClaims{}
					u.RegisteredClaims.ExpiresAt = &goJWT.NumericDate{Time: time.Now().Add(time.Minute).UTC()}
					return u, nil
				}())
				jwtCore.On(
					"GetTokenLoggedInDevices", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return("", errors.New("token not found"))
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message":"invalid token"}`,
		},
		{
			name:       "SUCCESS",
			reqSetting: addReqToken,
			mockSetup: func(jwtCore *jwt.IJwt) {
				jwtCore.On("GetTokenLoggedInDevices", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).Return("access_token", nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message": "OK"}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			test.reqSetting(req)
			test.mockSetup(mockJwt)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}

}
