package user_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/user"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnrollTOTP(t *testing.T) {
	service := serviceMocks.NewIUserService(t)

	handler := New(validatorExt.New(), service, nil, nil, nil, nil, nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Post("/enroll", handler.EnrollTOTP)

	userId := "faebe24f-5001-47a7-9899-4f75d1fb0fbf"
	userAuth := &model.UserTokenClaims{UUID: userId}

	tests := []struct {
		name             string
		userAuth         *model.UserTokenClaims
		requestBody      []byte
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR:User auth not found",
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR:Invalid request body",
			userAuth:         userAuth,
			requestBody:      []byte(`B`),
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid character 'B' looking for beginning of value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR:Invalid request data",
			userAuth:         userAuth,
			requestBody:      []byte(`{"qrCodeLevel": "TEST"}`),
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"QrCodeLevel","message":"Key: 'EnrollTOTPRequest.QrCodeLevel' Error:Field validation for 'QrCodeLevel' failed on the 'oneof' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR:Some error",
			userAuth:    userAuth,
			requestBody: []byte(`{}`),
			setupMock: func() {
				service.On(
					"EnrollTOTP", mock.Anything, mock.MatchedBy(func(req model.EnrollTOTPRequest) bool {
						return req.UserId == userId
					}),
				).Once().Return(nil, assert.AnError)
			},

			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS",
			userAuth:    userAuth,
			requestBody: []byte(`{"qrCodeLevel": "Medium", "qrCodeSize": 256, "UserId": "12345"}`),
			setupMock: func() {
				service.On(
					"EnrollTOTP", mock.Anything, mock.MatchedBy(func(req model.EnrollTOTPRequest) bool {
						return req.UserId == userId && req.QrCodeLevel == "Medium" && req.QrCodeSize == 256
					}),
				).Once().Return(&model.EnrollTOTPResponse{
					QRCodeDataURL: "data:image/png;base64,...",
					SecretKey:     "secret-key",
				}, nil)
			},

			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"qrCodeDataUrl":"data:image/png;base64,...","secretKey":"secret-key"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/enroll", bytes.NewReader(test.requestBody))

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userAuth != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userAuth))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Response Body:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}

func TestConfirmTOTP(t *testing.T) {
	service := serviceMocks.NewIUserService(t)

	handler := New(validatorExt.New(), service, nil, nil, nil, nil, nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Post("/confirm", handler.ConfirmTOTP)

	userId := "36551c79-2d8d-43f7-9c8d-366142b66387"
	userAuth := &model.UserTokenClaims{UUID: userId}

	tests := []struct {
		name             string
		userAuth         *model.UserTokenClaims
		requestBody      []byte
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR:User auth not found", // NOSONAR
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR:Invalid request body", // NOSONAR
			userAuth:         userAuth,
			requestBody:      []byte(`C`),
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"malformed request body payload","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR:Invalid request data", // NOSONAR
			userAuth:         userAuth,
			requestBody:      []byte(`{"otp": "123"}`),
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"OTP","message":"Key: 'ConfirmTOTPRequest.OTP' Error:Field validation for 'OTP' failed on the 'len' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR:Some error", // NOSONAR
			userAuth:    userAuth,
			requestBody: []byte(`{"otp": "123456"}`), // NOSONAR
			setupMock: func() {
				service.On("ConfirmTOTP", mock.Anything, mock.Anything).Once().Return(false, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS:Invalid OTP", // NOSONAR
			userAuth:    userAuth,
			requestBody: []byte(`{"otp": "123456"}`), // NOSONAR
			setupMock: func() {
				service.On("ConfirmTOTP", mock.Anything, mock.Anything).Once().Return(false, nil)
			},
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid TOTP code. please try again","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS:Valid OTP", // NOSONAR
			userAuth:    userAuth,
			requestBody: []byte(`{"otp": "123456"}`), // NOSONAR
			setupMock: func() {
				service.On("ConfirmTOTP", mock.Anything, mock.Anything).Once().Return(true, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"status":"ACTIVE"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/confirm", bytes.NewReader(test.requestBody))

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userAuth != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userAuth))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Response Body:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}

func TestSetPreferred2FAMethod(t *testing.T) {
	service := serviceMocks.NewIUserService(t)

	handler := New(validatorExt.New(), service, nil, nil, nil, nil, nil, nil, nil, nil)

	router := chi.NewRouter()
	router.Patch("/set-preferred-2fa", handler.SetPreferred2FAMethod)

	userId := "user-uuid-123"
	userAuth := &model.UserTokenClaims{UUID: userId}

	tests := []struct {
		name             string
		userAuth         *model.UserTokenClaims
		requestBody      []byte
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR:User auth not found",
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR:Invalid request body",
			userAuth:         userAuth,
			requestBody:      []byte(`D`),
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid character 'D' looking for beginning of value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR:Invalid request data - invalid method",
			userAuth:         userAuth,
			requestBody:      []byte(`{"preferred2FAMethod": "INVALID"}`),
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"Preferred2FAMethod","message":"Key: 'SetPreferred2FAMethodRequest.Preferred2FAMethod' Error:Field validation for 'Preferred2FAMethod' failed on the 'oneof' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR:Invalid request data - missing field",
			userAuth:         userAuth,
			requestBody:      []byte(`{}`),
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"Preferred2FAMethod","message":"Key: 'SetPreferred2FAMethodRequest.Preferred2FAMethod' Error:Field validation for 'Preferred2FAMethod' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR:Service error",
			userAuth:    userAuth,
			requestBody: []byte(`{"preferred2FAMethod": "OTP"}`),
			setupMock: func() {
				service.On("SetPreferred2FAMethod", mock.Anything, mock.MatchedBy(func(req model.SetPreferred2FAMethodRequest) bool {
					return req.UserId == userId && req.Preferred2FAMethod == "OTP"
				})).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: Set to OTP",
			userAuth:    userAuth,
			requestBody: []byte(`{"preferred2FAMethod": "OTP"}`),
			setupMock: func() {
				service.On("SetPreferred2FAMethod", mock.Anything, mock.MatchedBy(func(req model.SetPreferred2FAMethodRequest) bool {
					return req.UserId == userId && req.Preferred2FAMethod == "OTP"
				})).Once().Return(&model.SetPreferred2FAMethodResponse{
					Preferred2FAMethod: "OTP",
					Updated:            true,
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"preferred2FAMethod":"OTP","updated":true}}`,
		},
		{
			name:        "SUCCESS: Set to TOTP",
			userAuth:    userAuth,
			requestBody: []byte(`{"preferred2FAMethod": "TOTP"}`),
			setupMock: func() {
				service.On("SetPreferred2FAMethod", mock.Anything, mock.MatchedBy(func(req model.SetPreferred2FAMethodRequest) bool {
					return req.UserId == userId && req.Preferred2FAMethod == "TOTP"
				})).Once().Return(&model.SetPreferred2FAMethodResponse{
					Preferred2FAMethod: "TOTP",
					Updated:            true,
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"preferred2FAMethod":"TOTP","updated":true}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/set-preferred-2fa", bytes.NewReader(test.requestBody))

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userAuth != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userAuth))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Response Body:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
