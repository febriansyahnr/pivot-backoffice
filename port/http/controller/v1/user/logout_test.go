package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"
	mockUserRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/userRole"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserController_Logout(t *testing.T) {
	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	payloadRequest := &userModel.UserLogoutRequest{
		Email: "test@gmail.com",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name        string
		requestBody []byte
		mockSetup   func(
			userSvc *mockUser.IUserService,
			rabbitMqMock *mockRabbitMq.RabbitMQExt,
			jwtMock *mockJWT.IJwt)
		expectedStatus int
		userClaim      *userModel.UserTokenClaims
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(
				userSvc *mockUser.IUserService,
				rabbitMqMock *mockRabbitMq.RabbitMQExt,
				jwtMock *mockJWT.IJwt) {
				userSvc.
					On(
						"Logout",
						mock.Anything,
						mock.Anything).
					Return(nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)

				jwtMock.
					On(
						"RemoveIterateTokenFromRedis",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.Anything).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(
				userSvc *mockUser.IUserService,
				rabbitMqMock *mockRabbitMq.RabbitMQExt,
				jwtMock *mockJWT.IJwt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"email": "12345abcde"}`),
			mockSetup: func(userSvc *mockUser.IUserService,
				rabbitMqMock *mockRabbitMq.RabbitMQExt,
				jwtMock *mockJWT.IJwt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Service Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService,
				rabbitMqMock *mockRabbitMq.RabbitMQExt,
				jwtMock *mockJWT.IJwt) {
				userSvc.On("Logout",
					mock.Anything,
					mock.Anything,
				).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: User not in Context",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService,
				rabbitMqMock *mockRabbitMq.RabbitMQExt,
				jwtMock *mockJWT.IJwt) {
			},
			userClaim:      nil,
			expectedStatus: http.StatusUnauthorized,
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

			tt.mockSetup(mockUserSvc, mockRmq, jwtMock)

			mc := New(mockValidator, mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc, jwtMock, cfg, secret, mockRmq, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBuffer(tt.requestBody))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Logout)
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
