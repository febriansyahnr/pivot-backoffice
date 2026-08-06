package otp_test

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
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/otp"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVerifyOTP(pt *testing.T) {

	otpMock := serviceMocks.NewIOTP(pt)
	userMock := serviceMocks.NewIUserService(pt)

	router := chi.NewRouter()
	router.Post("/otp/verify", New(otpMock, userMock).Verify)

	tests := []struct {
		name           string
		reqBody        string
		mockSetup      func(otp *serviceMocks.IOTP)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			reqBody:        "B",
			mockSetup:      func(*serviceMocks.IOTP) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "invalid character 'B' looking for beginning of value"}`,
		},
		{
			name:           "ERROR:Data is invalid",
			reqBody:        `{"otp":""}`,
			mockSetup:      func(*serviceMocks.IOTP) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[{"field":"OTP", "message": "Key: 'VerifyOTPReq.OTP' Error:Field validation for 'OTP' failed on the 'required' tag"}],"traceId":"","type":"API_ERROR"},"message": "invalid validation"}`,
		},
		{
			name:    "ERROR:Password incorrect",
			reqBody: `{"otp":"123456"}`,
			mockSetup: func(otp *serviceMocks.IOTP) {
				otp.On(
					"ValidateOTPCode", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*otp.VerifyOTP"),
				).Once().Return("", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("password incorrect. please try again")))
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantRespBody:   `{"code": "45", "data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message": "password incorrect. please try again"}`,
		},
		{
			name:    "SUCCESS",
			reqBody: `{"otp":"123456"}`,
			mockSetup: func(otp *serviceMocks.IOTP) {
				otp.On(
					"ValidateOTPCode", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*otp.VerifyOTP"),
				).Return("<token-feature>", nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00", "message":"OK", "data": {"token": "<token-feature>", "isTOTPActive": false}}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/otp/verify", strings.NewReader(test.reqBody))
			req = req.WithContext(context.WithValue(req.Context(), constant.CtxTokenOTPKey, &otpModel.TokenOTPClaims{}))

			test.mockSetup(otpMock)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
