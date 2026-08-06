package user

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	chi "github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestForgotPassword(pt *testing.T) {

	userMock := serviceMocks.NewIUserService(pt)

	router := chi.NewRouter()
	router.Post(
		"/forgot-password", New(
			validator.New(), userMock, nil, nil, nil, nil, nil, nil, nil, nil,
		).ForgotPassword,
	)
	tests := []struct {
		name           string
		reqBody        string
		mockSetup      func(usr *serviceMocks.IUserService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			reqBody:        "B",
			mockSetup:      func(*serviceMocks.IUserService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "invalid character 'B' looking for beginning of value"}`,
		},
		{
			name:           "ERROR:Data is invalid",
			reqBody:        `{"email":"xxxx"}`,
			mockSetup:      func(*serviceMocks.IUserService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[{"field":"Email","message":"Key: 'UserForgotPasswordRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"}],"traceId":"","type":"API_ERROR"},"message": "invalid validation"}`,
		},
		{
			name:    "ERROR:Email not registered",
			reqBody: `{"email":"email@example.id"}`,
			mockSetup: func(usr *serviceMocks.IUserService) {
				usr.On(
					"ForgotPassword", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return("", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("email not registered")))
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantRespBody:   `{"code": "45", "data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message": "email not registered"}`,
		},
		{
			name:    "SUCCESS",
			reqBody: `{"email":"ready@example.id"}`,
			mockSetup: func(usr *serviceMocks.IUserService) {
				usr.On(
					"ForgotPassword", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return("<token-otp>", nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00","message":"OK", "data": {"token": "<token-otp>", "twoFactorAuthMethod":"", "isTOTPActive":false}}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(test.reqBody))

			test.mockSetup(userMock)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestResetPassword(pt *testing.T) {

	userMock := serviceMocks.NewIUserService(pt)

	router := chi.NewRouter()
	router.Patch(
		"/reset-password", New(
			validator.New(), userMock, nil, nil, nil, nil, nil, nil, nil, nil,
		).ResetPassword,
	)
	tests := []struct {
		name           string
		reqBody        string
		mockSetup      func(usr *serviceMocks.IUserService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			reqBody:        "B",
			mockSetup:      func(*serviceMocks.IUserService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "invalid character 'B' looking for beginning of value"}`,
		},
		{
			name:           "ERROR:Data is invalid",
			reqBody:        `{"password":"123"}`,
			mockSetup:      func(*serviceMocks.IUserService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[{"field":"Password","message":"Key: 'UserResetPasswordRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag"}],"traceId":"","type":"API_ERROR"},"message": "invalid validation"}`,
		},
		{
			name:    "ERROR:Invalid session",
			reqBody: `{"password":"12345678h"}`,
			mockSetup: func(usr *serviceMocks.IUserService) {
				usr.On(
					"ResetPassword", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(pkgErrs.New(response.HttpErrDatabase, errors.New("invalid session")))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code": "98", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "invalid session"}`,
		},
		{
			name:    "SUCCESS",
			reqBody: `{"password":"12345678h"}`,
			mockSetup: func(usr *serviceMocks.IUserService) {
				usr.On(
					"ResetPassword", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00", "message":"OK","data": {"updated": true}}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/reset-password", strings.NewReader(test.reqBody))

			req = req.WithContext(context.WithValue(req.Context(), constant.CtxTokenOTPKey, &otpModel.TokenOTPClaims{}))

			test.mockSetup(userMock)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
