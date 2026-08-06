package user

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
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

func TestUserController_AddUserRole(t *testing.T) {
	now := time.Now()

	expectedUser := &userModel.User{
		UUID:       "user-id",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "ganteng123",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  now,
	}

	expectedRole := &role.Role{
		UUID:      "role-id",
		Name:      "test",
		Slug:      "test",
		CreatedAt: now,
	}

	//expectedUserRole := &userRole.UserRole{
	//	UUID:      "uuid-uuid-uuid",
	//	UserID:    "user-id",
	//	RoleID:    "role-id",
	//	CreatedAt: now,
	//}

	paylod := &userModel.UserAddRoleRequest{
		Email:    "test@gmail.com",
		RoleSlug: "role-id",
	}

	payloadRequestByte, err := json.Marshal(paylod)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		expectedUser   *userModel.User
		expectedRole   *role.Role
		requestBody    []byte
		mockSetup      func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRoleSvc *mockUserRole.IUserRoleService)
		expectedStatus int
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRoleSvc *mockUserRole.IUserRoleService) {
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(expectedUser, nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(expectedRole, nil)
				userRoleSvc.
					On(
						"Create",
						mock.Anything,
						mock.Anything).
					Return(nil)
			},
			expectedStatus: 200,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRoleSvc *mockUserRole.IUserRoleService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"email": "12345abcde"}`),
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRoleSvc *mockUserRole.IUserRoleService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR: Service Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRoleSvc *mockUserRole.IUserRoleService) {
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(expectedUser, nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(expectedRole, nil)
				userRoleSvc.
					On(
						"Create",
						mock.Anything,
						mock.Anything).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "ERROR: user not found",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRoleSvc *mockUserRole.IUserRoleService) {
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "ERROR: role not found",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRoleSvc *mockUserRole.IUserRoleService) {
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(expectedUser, nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "ERROR: failed to find user",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRoleSvc *mockUserRole.IUserRoleService) {
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "ERROR: failed to find role",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRoleSvc *mockUserRole.IUserRoleService) {
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(expectedUser, nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Return(nil, errors.New("service error"))
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

			tt.mockSetup(mockUserSvc, mockRoleSvc, userRoleMock)

			mc := New(mockValidator, mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc, jwtMock, cfg, secret, mockRmq, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/roles/assign", bytes.NewBuffer(tt.requestBody))
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.AddUserRole)
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
