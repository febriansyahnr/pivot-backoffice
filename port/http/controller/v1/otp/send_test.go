package otp_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/otp"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSend(pt *testing.T) {
	otpMock := serviceMocks.NewIOTP(pt)
	userMock := serviceMocks.NewIUserService(pt)

	router := chi.NewRouter()
	router.Post("/otp/send", New(otpMock, userMock).Send)

	tests := []struct {
		name           string
		reqBody        string
		mockSetup      func(otp *serviceMocks.IOTP, user *serviceMocks.IUserService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			reqBody:        "X",
			mockSetup:      func(*serviceMocks.IOTP, *serviceMocks.IUserService) { /* Empty Function */ },
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "invalid character 'X' looking for beginning of value"}`,
		},
		{
			name:           "ERROR:Data is invalid",
			reqBody:        `{"email": "", "event": "reset-pin"}`,
			mockSetup:      func(*serviceMocks.IOTP, *serviceMocks.IUserService) { /* Empty Function */ },
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[{"field":"Email", "message":"Key: 'SendOTPReq.Email' Error:Field validation for 'Email' failed on the 'required' tag"}],"traceId":"","type":"API_ERROR"},"message": "invalid validation"}`,
		},
		{
			name:    "ERROR:Email not registered",
			reqBody: `{"email": "dummy@random.com", "event": "reset-pin"}`,
			mockSetup: func(otp *serviceMocks.IOTP, user *serviceMocks.IUserService) {
				otp.On(
					"SendGenerateOTPCode", mock.Anything, mock.Anything,
				).Once().Return("", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("email not registered")))
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantRespBody:   `{"code": "45", "data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message": "email not registered"}`,
		},
		{
			name:    "SUCCESS: 2FA method OTP - TOTP not active",
			reqBody: `{"email": "dummy@random.com", "event": "reset-pin"}`,
			mockSetup: func(otp *serviceMocks.IOTP, user *serviceMocks.IUserService) {
				otp.On("SendGenerateOTPCode", mock.Anything, mock.Anything).Once().Return("<token-otp>", nil)
				user.On("FindUserByEmail", mock.Anything, "dummy@random.com").Once().Return(&userModel.User{UUID: "user-uuid", TOTPStatus: constant.TOTPStatusNotEnrolled}, nil)
				user.On("FindUserTOTPDataByID", mock.Anything, "user-uuid").Once().Return(&userModel.UserTOTPData{TOTPStatus: constant.TOTPStatusNotEnrolled}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00", "message":"OK", "data": {"token": "<token-otp>","twoFactorAuthMethod":"OTP","isTOTPActive":false}}`,
		},
		{
			name:    "SUCCESS: 2FA method OTP - TOTP active",
			reqBody: `{"email": "active-totp@random.com", "event": "reset-pin"}`,
			mockSetup: func(otp *serviceMocks.IOTP, user *serviceMocks.IUserService) {
				otp.On("SendGenerateOTPCode", mock.Anything, mock.Anything).Once().Return("<token-otp>", nil)
				user.On("FindUserByEmail", mock.Anything, "active-totp@random.com").Once().Return(&userModel.User{UUID: "user-uuid-2", TOTPStatus: constant.TOTPStatusActive}, nil)
				user.On("FindUserTOTPDataByID", mock.Anything, "user-uuid-2").Once().Return(&userModel.UserTOTPData{TOTPStatus: constant.TOTPStatusActive}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00", "message":"OK", "data": {"token": "<token-otp>","twoFactorAuthMethod":"OTP","isTOTPActive":true}}`,
		},
		{
			name:    "SUCCESS: 2FA method TOTP",
			reqBody: `{"email": "dummy@random.com", "event": "change-password"}`,
			mockSetup: func(otp *serviceMocks.IOTP, user *serviceMocks.IUserService) {
				otp.On("SendGenerateOTPCode", mock.Anything, mock.Anything).Once().Return(constant.TOTPTokenPrefixID+"<token-otp>", nil)
				user.On("FindUserByEmail", mock.Anything, "dummy@random.com").Once().Return(&userModel.User{UUID: "user-uuid", TOTPStatus: constant.TOTPStatusActive}, nil)
				user.On("FindUserTOTPDataByID", mock.Anything, "user-uuid").Once().Return(&userModel.UserTOTPData{TOTPStatus: constant.TOTPStatusActive}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00", "message":"OK", "data": {"token": "<token-otp>","twoFactorAuthMethod":"TOTP","isTOTPActive":true}}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/otp/send", strings.NewReader(test.reqBody))

			test.mockSetup(otpMock, userMock)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
