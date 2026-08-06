package user_test

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
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/user"

	chi "github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResetPIN(pt *testing.T) {

	userMock := serviceMocks.NewIUserService(pt)

	router := chi.NewRouter()
	router.Patch(
		"/reset-pin", New(
			validator.New(), userMock, nil, nil, nil, nil, nil, nil, nil, nil,
		).ResetPIN,
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
			reqBody:        `{"pin":""}`,
			mockSetup:      func(*serviceMocks.IUserService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[{"field":"Pin","message":"Key: 'ResetPinRequest.Pin' Error:Field validation for 'Pin' failed on the 'required' tag"}],"traceId":"","type":"API_ERROR"},"message": "invalid validation"}`,
		},
		{
			name:    "ERROR:Invalid session",
			reqBody: `{"pin":"123456"}`,
			mockSetup: func(usr *serviceMocks.IUserService) {
				usr.On(
					"ResetPIN", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(pkgErrs.New(response.HttpErrDatabase, errors.New("invalid session")))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code": "98", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "invalid session"}`,
		},
		{
			name:    "SUCCESS",
			reqBody: `{"pin":"123456"}`,
			mockSetup: func(usr *serviceMocks.IUserService) {
				usr.On(
					"ResetPIN", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00", "message":"OK", "data": {"updated": true}}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/reset-pin", strings.NewReader(test.reqBody))

			req = req.WithContext(context.WithValue(req.Context(), constant.CtxTokenOTPKey, &otpModel.TokenOTPClaims{}))

			test.mockSetup(userMock)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
