package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
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

func TestUserController_Register(t *testing.T) {
	expectedRole := &role.Role{
		UUID:      uuid.NewString(),
		Name:      "admin",
		Slug:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	payloadRequest := &userModel.UserRegisterRequest{
		Email:                "test@gmail.com",
		Name:                 "test",
		Password:             "ganteng123",
		PasswordConfirmation: "ganteng123",
		CreatedAt:            time.Now(),
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		expectedRole   *role.Role
		mockSetup      func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt)
		expectedStatus int
	}{
		{
			name:         "SUCCESS",
			requestBody:  payloadRequestByte,
			expectedRole: expectedRole,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
				userSvc.
					On(
						"Create",
						mock.Anything,
						mock.Anything).
					Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(expectedRole, nil)
				userRole.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*userRole.UserRole")).
					Return(nil)
				mockJwt.
					On(
						"GenerateRefreshToken",
						mock.Anything,
						mock.Anything,
						mock.AnythingOfType(constant.MockTypeTime)).
					Return(mock.Anything, nil)
				userSvc.
					On(
						"Update",
						mock.Anything,
						mock.Anything).
					Return(nil)
			},
			expectedStatus: 200,
		},
		{
			name:        "ERROR: Failed to generate refresh token",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
				userSvc.
					On(
						"Create",
						mock.Anything,
						mock.Anything).
					Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(expectedRole, nil)
				userRole.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*userRole.UserRole")).
					Return(nil)
				mockJwt.
					On(
						"GenerateRefreshToken",
						mock.Anything,
						mock.Anything,
						mock.AnythingOfType(constant.MockTypeTime)).
					Return("", errors.New("failed to generate refresh token"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "ERROR: Failed to update user",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
				userSvc.
					On(
						"Create",
						mock.Anything,
						mock.Anything).
					Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(expectedRole, nil)
				userRole.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*userRole.UserRole")).
					Return(nil)
				mockJwt.
					On(
						"GenerateRefreshToken",
						mock.Anything,
						mock.Anything,
						mock.AnythingOfType(constant.MockTypeTime)).
					Return(mock.Anything, nil)
				userSvc.
					On(
						"Update",
						mock.Anything,
						mock.Anything).
					Return(errors.New("failed to update user"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"email": "12345abcde"}`),
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "ERROR: Failed to get role",
			requestBody:  payloadRequestByte,
			expectedRole: nil,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
				userSvc.
					On(
						"Create",
						mock.Anything,
						mock.Anything).
					Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(nil, errors.New("failed to get role"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:         "ERROR: Role not found",
			requestBody:  payloadRequestByte,
			expectedRole: nil,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
				userSvc.
					On(
						"Create",
						mock.Anything,
						mock.Anything).
					Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(nil, errors.New("not found"))
				roleSvc.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*role.Role")).
					Return(nil, nil)
				userRole.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*userRole.UserRole")).
					Return(nil)
				mockJwt.
					On(
						"GenerateRefreshToken",
						mock.Anything,
						mock.Anything,
						mock.AnythingOfType(constant.MockTypeTime)).
					Return(mock.Anything, nil)
				userSvc.
					On(
						"Update",
						mock.Anything,
						mock.Anything).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "ERROR: Failed to create role",
			requestBody:  payloadRequestByte,
			expectedRole: nil,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
				userSvc.
					On(
						"Create",
						mock.Anything,
						mock.Anything).
					Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(nil, errors.New("role not found"))
				roleSvc.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*role.Role")).
					Return(errors.New("failed to create role"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:         "ERROR: Failed to create user role",
			requestBody:  payloadRequestByte,
			expectedRole: expectedRole,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
				userSvc.
					On(
						"Create",
						mock.Anything,
						mock.Anything).
					Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(expectedRole, nil)
				userRole.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*userRole.UserRole")).
					Return(errors.New("failed to create user role"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "ERROR: Service Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, mockJwt *mockJWT.IJwt) {
				userSvc.On("Create",
					mock.Anything,
					mock.Anything,
				).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
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

			tt.mockSetup(mockUserSvc, mockRoleSvc, userRoleMock, jwtMock)

			mc := New(mockValidator, mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc, jwtMock, cfg, secret, mockRmq, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/users/create", bytes.NewBuffer(tt.requestBody))
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Register)
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
