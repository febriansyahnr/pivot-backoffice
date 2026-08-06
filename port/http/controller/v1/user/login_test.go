package user_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"
	mockUserRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/userRole"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/user"

	chi "github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserController_Login(t *testing.T) {
	expectedUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "ganteng123",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  time.Now(),
	}

	payloadRequest := &userModel.UserLoginRequest{
		Email:    "test@gmail.com",
		Password: "ganteng123",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				userSvc.
					On(
						"Login",
						mock.Anything,
						mock.Anything,
						mock.Anything).
					Return(expectedUser, "token", nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: 200,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"email": "12345abcde"}`),
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR: Service Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				userSvc.On("Login",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, "", errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "ERROR: user not found",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				userSvc.On("Login",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, "", nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "SUCCESS: 2FA required - TOTP not active",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				userSvc.On("Login",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, "", constant.ErrNeed2FAChallengeForLogin)
				userSvc.On("LoginWithOTP",
					mock.Anything,
					"test@gmail.com",
					"ganteng123",
				).Return("<otp-token>", nil)
				userSvc.On("FindUserByEmail",
					mock.Anything,
					"test@gmail.com",
				).Return(&userModel.User{UUID: "user-uuid", TOTPStatus: constant.TOTPStatusNotEnrolled}, nil)
				userSvc.On("FindUserTOTPDataByID",
					mock.Anything,
					"user-uuid",
				).Return(&userModel.UserTOTPData{TOTPStatus: constant.TOTPStatusNotEnrolled}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "SUCCESS: 2FA required - TOTP active",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				userSvc.On("Login",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, "", constant.ErrNeed2FAChallengeForLogin)
				userSvc.On("LoginWithOTP",
					mock.Anything,
					"test@gmail.com",
					"ganteng123",
				).Return(constant.TOTPTokenPrefixID+"<totp-token>", nil)
				userSvc.On("FindUserByEmail",
					mock.Anything,
					"test@gmail.com",
				).Return(&userModel.User{UUID: "user-uuid", TOTPStatus: constant.TOTPStatusActive}, nil)
				userSvc.On("FindUserTOTPDataByID",
					mock.Anything,
					"user-uuid",
				).Return(&userModel.UserTOTPData{TOTPStatus: constant.TOTPStatusActive}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					UserKey: "testing",
				},
			}

			mockUserSvc := mockUser.NewIUserService(t)
			mockRoleSvc := mockRole.NewIRoleService(t)
			userRoleMock := mockUserRole.NewIUserRoleService(t)
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			jwtMock := mockJWT.NewIJwt(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tt.mockSetup(mockUserSvc, mockRmq)

			mc := New(mockValidator, mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc, jwtMock, cfg, secret, mockRmq, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(tt.requestBody))
			req.Header.Set(constant.HeaderDeviceIdentifier, "device-id")
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Login)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockUserSvc.AssertExpectations(t)
		})
	}
}

func TestLoginWithOTP(pt *testing.T) {
	userMock := mockUser.NewIUserService(pt)

	router := chi.NewRouter()
	router.Post(
		"/auth/login", New(
			validator.New(), userMock, nil, nil, nil, nil, nil, nil, nil, nil,
		).LoginWithOTP,
	)
	tests := []struct {
		name           string
		reqBody        string
		mockSetup      func(usr *mockUser.IUserService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			reqBody:        "B",
			mockSetup:      func(*mockUser.IUserService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "invalid character 'B' looking for beginning of value"}`,
		},
		{
			name:           "ERROR:Data is invalid",
			reqBody:        `{"email":"xxxx","password":"123"}`,
			mockSetup:      func(*mockUser.IUserService) {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code": "40", "data":null,"error":{"details":[{"field":"Email","message":"Key: 'UserLoginRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"}],"traceId":"","type":"API_ERROR"},"message": "invalid validation"}`,
		},
		{
			name:    "ERROR:Email not registered",
			reqBody: `{"email":"unregistered@example.id","password":"123"}`, // NOSONAR
			mockSetup: func(u *mockUser.IUserService) {
				u.On(
					"LoginWithOTP", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything,
				).Once().Return("", pkgErrs.New(response.HttpErrUnauthorized, errors.New("incorrect email or password.")))
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code": "41", "data":null,"error":{"details":[],"traceId":"","type":"API_ERROR"},"message": "incorrect email or password."}`,
		},
		{
			name:    "SUCCESS:2FA with OTP - TOTP not active",             // NOSONAR
			reqBody: `{"email":"registered@example.id","password":"123"}`, // NOSONAR
			mockSetup: func(u *mockUser.IUserService) {
				u.On(
					"LoginWithOTP", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything,
				).Once().Return("<token-otp>", nil)
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), "registered@example.id",
				).Once().Return(&userModel.User{UUID: "user-uuid", TOTPStatus: constant.TOTPStatusNotEnrolled}, nil)
				u.On(
					"FindUserTOTPDataByID", mock.AnythingOfType(constant.MockTypeValueContextReference), "user-uuid",
				).Once().Return(&userModel.UserTOTPData{TOTPStatus: constant.TOTPStatusNotEnrolled}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00", "message":"OK", "data": {"token": "<token-otp>","twoFactorAuthMethod":"OTP","isTOTPActive":false}}`,
		},
		{
			name:    "SUCCESS:2FA with OTP - TOTP active",                 // NOSONAR
			reqBody: `{"email":"registered@example.id","password":"123"}`, // NOSONAR
			mockSetup: func(u *mockUser.IUserService) {
				u.On(
					"LoginWithOTP", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything,
				).Once().Return("<token-otp>", nil)
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), "registered@example.id",
				).Once().Return(&userModel.User{UUID: "user-uuid", TOTPStatus: constant.TOTPStatusActive}, nil)
				u.On(
					"FindUserTOTPDataByID", mock.AnythingOfType(constant.MockTypeValueContextReference), "user-uuid",
				).Once().Return(&userModel.UserTOTPData{TOTPStatus: constant.TOTPStatusActive}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00", "message":"OK", "data": {"token": "<token-otp>","twoFactorAuthMethod":"OTP","isTOTPActive":true}}`,
		},
		{
			name:    "SUCCESS:2FA with TOTP",                              // NOSONAR
			reqBody: `{"email":"registered@example.id","password":"123"}`, // NOSONAR
			mockSetup: func(u *mockUser.IUserService) {
				u.On(
					"LoginWithOTP", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything,
				).Once().Return(constant.TOTPTokenPrefixID+"<token-otp>", nil)
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), "registered@example.id",
				).Once().Return(&userModel.User{UUID: "user-uuid", TOTPStatus: constant.TOTPStatusActive}, nil)
				u.On(
					"FindUserTOTPDataByID", mock.AnythingOfType(constant.MockTypeValueContextReference), "user-uuid",
				).Once().Return(&userModel.UserTOTPData{TOTPStatus: constant.TOTPStatusActive}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00", "message":"OK", "data": {"token": "<token-otp>","twoFactorAuthMethod":"TOTP","isTOTPActive":true}}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(test.reqBody))

			test.mockSetup(userMock)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestSessionFromLogin2FA(pt *testing.T) {
	userMock := mockUser.NewIUserService(pt)

	router := chi.NewRouter()
	router.Get(
		"/users/2fa/token", New(
			validator.New(), userMock, nil, nil, nil, nil, nil, nil, nil, nil,
		).SessionFromLogin2FA,
	)
	tests := []struct {
		name           string
		mockSetup      func(u *mockUser.IUserService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR:Generate Token",
			mockSetup: func(u *mockUser.IUserService) {
				u.On(
					"GenerateTokenFromLogin2FA", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything,
				).Once().Return(nil, "", errors.New("invalid redis session"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code": "99", "data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message": "invalid redis session"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func(u *mockUser.IUserService) {
				u.On(
					"GenerateTokenFromLogin2FA", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything,
				).Once().Return(&userModel.User{UUID: "unique-id", Email: "registered@example.id"}, "access-token", nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"userInfo":{"uuid":"unique-id","email":"registered@example.id","status":"","name":"","blockedAt":"0001-01-01T00:00:00Z","merchantId":"","isChangePassword":0,"isEmptyPin":1,"role":"","totpStatus":"","preferred2FAMethod":"","deactivatedAt":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"},"accessToken":"access-token","refreshToken":""}}`,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/users/2fa/token", nil)

			req = req.WithContext(context.WithValue(req.Context(), constant.CtxTokenOTPKey, &otpModel.TokenOTPClaims{}))
			test.mockSetup(userMock)

			router.ServeHTTP(rec, req)
			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
