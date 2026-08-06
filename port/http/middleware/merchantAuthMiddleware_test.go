package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	jwt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware"

	chi "github.com/go-chi/chi/v5"
	goJWT "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMerchantAuthMiddleware(pt *testing.T) {
	mockJwt := jwt.NewIJwt(pt)
	router := chi.NewRouter()
	MountHandlers(router, middleware.MerchantAuthMiddleware(mockJwt))

	addReqToken := func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer dummy-token")
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
			wantRespBody:   `{"code":"credentials_invalid","error":{"details":[{"field":"","message":"Request new access token"}],"traceId":"","type":"API_ERROR"},"message":"Access token is invalid"}`,
		},
		{
			name:       "ERROR:Token not recognized",
			reqSetting: addReqToken,
			mockSetup: func(jwtCore *jwt.IJwt) {
				jwtCore.On(
					"VerifyMerchantToken", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, errors.New("token not recognized"))
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"credentials_invalid","error":{"details":[{"field":"","message":"Request new access token"}],"traceId":"","type":"API_ERROR"},"message":"Access token is invalid"}`,
		},
		{
			name:       "ERROR:Token has expired",
			reqSetting: addReqToken,
			mockSetup: func(jwtCore *jwt.IJwt) {
				jwtCore.On(
					"VerifyMerchantToken", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(func() (*merchant.MerchantAuthTokenClaims, error) {
					u := &merchant.MerchantAuthTokenClaims{}
					u.RegisteredClaims.ExpiresAt = &goJWT.NumericDate{Time: time.Now().Add(-time.Minute).UTC()}
					return u, nil
				}())
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name:       "SUCCESS",
			reqSetting: addReqToken,
			mockSetup: func(jwtCore *jwt.IJwt) {
				jwtCore.On(
					"VerifyMerchantToken", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(func() (*merchant.MerchantAuthTokenClaims, error) {
					u := &merchant.MerchantAuthTokenClaims{}
					u.RegisteredClaims.ExpiresAt = &goJWT.NumericDate{Time: time.Now().Add(time.Minute).UTC()}
					return u, nil
				}())
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
