package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	jwtPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	jwtExt "github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"

	chi "github.com/go-chi/chi/v5"
	redismock "github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthTokenOTP(pt *testing.T) {
	emailExample := "email@example.id"
	authToken := "486fa420-4983-4449-92a5-e2e8cc88e1dd"
	verifyOtpToken := fmt.Sprintf(`{"token": "%s"}`, authToken)

	userActivatePath := "/users/activate"
	resetPasswordSuccessPath := "/auth/reset-password/success"
	resetPasswordFailedPath := "/auth/reset-password/failed"

	for i, middleware := range []func(jwtExt.IJwt, redisExt.IRedisExt) MiddlewareFunc{AuthTokenFromOTP, AuthTokenFromFeature2FA} {

		jwtMock := jwtPkgMock.NewIJwt(pt)

		db, clientMock := redismock.NewClientMock()
		redisMock := redisExt.WrapRedisClient(db, nil)

		router := chi.NewRouter()
		router.Use(middleware(jwtMock, redisMock))
		router.Post(resetPasswordSuccessPath, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"message": "OK"}`))
			w.WriteHeader(http.StatusOK)
		})
		router.Post(resetPasswordFailedPath, func(wr http.ResponseWriter, _ *http.Request) {
			wr.WriteHeader(http.StatusBadRequest)
			_, _ = wr.Write([]byte(`{"code":"40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "password incorrect"}`))
		})
		router.Post(userActivatePath, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"message": "OK"}`))
			w.WriteHeader(http.StatusOK)
		})

		funcMock := []string{"ValidateTokenFromOTP", "ValidateTokenFromFeature2FA"}[i]
		authKey := []string{
			"backend-portal:otp-verification:UUID:forgot-password:token-otp",
			"backend-portal:otp-verification:UUID:forgot-password:token-feature:" + authToken,
		}[i]
		cacheValue := []string{verifyOtpToken, emailExample}[i]

		tests := []struct {
			name           string
			token          string
			url            string
			reqSetup       func(token string, r *http.Request)
			mockSetup      func(j *jwtPkgMock.IJwt, r redismock.ClientMock)
			wantStatusCode int
			wantRespBody   string
		}{
			{
				name:           "ERROR:Token Not Found",
				wantStatusCode: http.StatusUnauthorized,
				wantRespBody:   `{"code": "41", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "invalid token format"}`,
			},
			{
				name: "ERROR:Token Expired",
				reqSetup: func(token string, r *http.Request) {
					r.Header.Set(constant.HeaderAuthorization, "Bearer "+token)
				},
				mockSetup: func(j *jwtPkgMock.IJwt, _ redismock.ClientMock) {
					j.On(
						funcMock, mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything,
					).Once().Return(nil, errors.New("token has invalid claims: token is expired"))
				},
				wantStatusCode: http.StatusUnauthorized,
				wantRespBody:   `{"code": "41", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "token has invalid claims: token is expired"}`,
			},
			{
				name:  "MULTI:Error If Feature Authorization",
				token: authToken,
				url:   userActivatePath,
				reqSetup: func(token string, r *http.Request) {
					r.Header.Set(constant.HeaderAuthorization, "Bearer "+token)
				},
				mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock) {
					j.On(
						funcMock, mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
					).Return(&otp.TokenOTPClaims{UUID: "UUID", Identifier: constant.OTPIdentifierForgotPassword}, nil)

					r.ExpectGet(authKey).SetVal(cacheValue)
				},
				wantStatusCode: []int{http.StatusOK, http.StatusUnauthorized}[i],
				wantRespBody:   []string{`{"message": "OK"}`, `{"code":"41", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "invalid feature token"}`}[i],
			},
			{
				name:  "ERROR:Token is not registered",
				token: authToken,
				reqSetup: func(token string, r *http.Request) {
					r.Header.Set(constant.HeaderAuthorization, "Bearer "+token)
				},
				mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock) {
					if funcMock == "ValidateTokenFromOTP" {
						r.ExpectGet(authKey).SetVal(`{"token":"xxx"}`)
					} else {
						r.ExpectGet(authKey).RedisNil()
					}
				},
				wantStatusCode: http.StatusUnauthorized,
				wantRespBody:   `{"code": "41", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "token is not registered"}`,
			},
			{
				name:  "SUCCESS:Invoked Reset Password Failed",
				token: authToken,
				url:   resetPasswordFailedPath,
				reqSetup: func(token string, r *http.Request) {
					r.Header.Set(constant.HeaderAuthorization, "Bearer "+token)
				},
				mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock) {
					r.ExpectGet(authKey).SetVal(cacheValue)
				},
				wantStatusCode: http.StatusBadRequest,
				wantRespBody:   `{"code":"40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "password incorrect"}`,
			},
			{
				name:  "SUCCESS:Invoked Reset Password Success",
				token: authToken,
				reqSetup: func(token string, r *http.Request) {
					r.Header.Set(constant.HeaderAuthorization, "Bearer "+token)
				},
				mockSetup: func(j *jwtPkgMock.IJwt, r redismock.ClientMock) {
					r.ExpectGet(authKey).SetVal(cacheValue)
				},
				wantStatusCode: http.StatusOK,
				wantRespBody:   `{"message":"OK"}`,
			},
		}
		for _, test := range tests {
			pt.Run("["+funcMock+"]"+test.name, func(t *testing.T) {

				clientMock.ClearExpect()

				if test.url == "" {
					test.url = resetPasswordSuccessPath
				}
				if test.mockSetup != nil {
					test.mockSetup(jwtMock, clientMock)
				}
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, test.url, nil)

				if test.reqSetup != nil {
					test.reqSetup(test.token, req)
				}

				router.ServeHTTP(rec, req)
				require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
				assert.JSONEq(t, test.wantRespBody, rec.Body.String())
			})
		}
	}
}

func TestSpecialCaseRequireAuthForSendOTP(pt *testing.T) {
	jwtMock := jwtPkgMock.NewIJwt(pt)

	route := chi.NewRouter()
	route.Use(SpecialCaseRequireAuthForSendOTP(jwtMock))
	route.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message": "OK"}`))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	const token = "example-token"

	tests := []struct {
		name           string
		reqBody        string
		mockSetup      func(j *jwtPkgMock.IJwt)
		reqSetup       func(r *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			reqBody:        ``,
			mockSetup:      func(j *jwtPkgMock.IJwt) {},
			reqSetup:       func(r *http.Request) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "unexpected end of JSON input"}`,
		},
		{
			name:           "ERROR:JSON is malformed",
			reqBody:        `{"email":{}}`,
			mockSetup:      func(j *jwtPkgMock.IJwt) {},
			reqSetup:       func(r *http.Request) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "json: cannot unmarshal object into Go struct field jsonSendOTPReq.email of type string"}`,
		},
		{
			name:           "SUCCESS:Non auth event",
			reqBody:        `{"email": "john.wick@example.id", "event": "forgot-password"}`,
			mockSetup:      func(j *jwtPkgMock.IJwt) {},
			reqSetup:       func(r *http.Request) {},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message": "OK"}`,
		},
		{
			name:           "ERROR:Token required on reset pin",
			reqBody:        `{"email": "john.wick@example.id", "event": "reset-pin"}`,
			mockSetup:      func(j *jwtPkgMock.IJwt) {},
			reqSetup:       func(r *http.Request) {},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code": "41", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "token required"}`,
		},
		{
			name:           "ERROR:Token required on change password",
			reqBody:        `{"email": "john.wick@example.id", "event": "change-password"}`,
			mockSetup:      func(j *jwtPkgMock.IJwt) {},
			reqSetup:       func(r *http.Request) {},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code": "41", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "token required"}`,
		},
		{
			name:    "ERROR:Verify token",
			reqBody: `{"email": "john.wick@example.id", "event": "change-password"}`,
			mockSetup: func(j *jwtPkgMock.IJwt) {
				j.On(
					"Verify", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, errors.New("token expired"))
			},
			reqSetup: func(r *http.Request) {
				r.Header.Set("X-Access-Token", token)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code": "41", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "token expired"}`,
		},
		{
			name:    "ERROR:Email is not associated with account",
			reqBody: `{"email": "widya.bagus@example.id", "event": "reset-pin"}`,
			mockSetup: func(j *jwtPkgMock.IJwt) {
				j.On(
					"Verify", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(&user.UserTokenClaims{Email: "john.wick@example.id"}, nil)
			},
			reqSetup: func(r *http.Request) {
				r.Header.Set("X-Access-Token", token)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "this email is not associated with your account"}`,
		},
		{
			name:    "ERROR:Token not registred",
			reqBody: `{"email": "john.wick@example.id", "event": "reset-pin"}`,
			mockSetup: func(j *jwtPkgMock.IJwt) {
				j.On(
					"Verify", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&user.UserTokenClaims{Email: "john.wick@example.id"}, nil)

				j.On(
					"GetTokenLoggedInDevices", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return("", errors.New("redis nil"))
			},
			reqSetup: func(r *http.Request) {
				r.Header.Set("X-Access-Token", token)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code": "41", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "token not registered"}`,
		},
		{
			name:    "SUCCESS:Auth event",
			reqBody: `{"email": "john.wick@example.id", "event": "reset-pin"}`,
			mockSetup: func(j *jwtPkgMock.IJwt) {
				j.On(
					"GetTokenLoggedInDevices", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return("", nil)
			},
			reqSetup: func(r *http.Request) {
				r.Header.Set("X-Access-Token", token)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"message": "OK"}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader(test.reqBody))

			test.reqSetup(req)
			test.mockSetup(jwtMock)

			route.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
