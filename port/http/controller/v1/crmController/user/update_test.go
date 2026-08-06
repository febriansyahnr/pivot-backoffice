package user

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	service "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRMUserController_Update(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        []byte
		userID             string
		setupMocks         func(*service.IMerchantService, *service.IUserService, *service.IRoleService, *service.IUserRoleService, *mockRabbitMq.RabbitMQExt)
		expectedStatusCode int
	}{
		{
			name:        "Success",
			userID:      "user-123",
			requestBody: []byte(`{"name": "John Doe", "email": "john.doe@example.com", "roleSlug": "admin", "status": "BLOCKED"}`),
			setupMocks: func(ms *service.IMerchantService, us *service.IUserService, rs *service.IRoleService, urs *service.IUserRoleService, rabbitMqExt *mockRabbitMq.RabbitMQExt) {
				us.On("FindUserByID", mock.Anything, "user-123").Return(&user.User{
					UUID:       "user-123",
					MerchantId: "merchant-123",
				}, nil)
				us.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
				rs.On("FindRoleBySlug", mock.Anything, "admin").Return(&role.Role{UUID: "role-123", Slug: "admin"}, nil)
				urs.On("FindUserRoleByUserID", mock.Anything, "user-123").Return(&userRole.UserRole{}, nil)
				urs.On("UpdateByUserID", mock.Anything, mock.AnythingOfType("*userRole.UserRole")).Return(nil)
				rabbitMqExt.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "Invalid User ID",
			requestBody: []byte(`{"name": "John Doe", "email": "john.doe@example.com", "roleSlug": "admin", "status": "BLOCKED"}`),
			userID:      "",
			setupMocks: func(ms *service.IMerchantService, us *service.IUserService, rs *service.IRoleService, urs *service.IUserRoleService, rabbitMqExt *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "User Not Found",
			requestBody: []byte(`{"name": "John Doe", "email": "john.doe@example.com", "roleSlug": "admin", "status": "BLOCKED"}`),
			userID:      "user-123",
			setupMocks: func(ms *service.IMerchantService, us *service.IUserService, rs *service.IRoleService, urs *service.IUserRoleService, rabbitMqExt *mockRabbitMq.RabbitMQExt) {
				us.On("FindUserByID", mock.Anything, "user-123").Return(nil, errors.New("user not found"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "Invalid JSON Payload",
			requestBody: []byte(`invalid json`),
			userID:      "user-123",
			setupMocks: func(ms *service.IMerchantService, us *service.IUserService, rs *service.IRoleService, urs *service.IUserRoleService, rabbitMqExt *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "ERROR: validate struct payload",
			userID:      "user-123",
			requestBody: []byte(`{"name": "John Doe", "email": "john.doe@example.com"}`),
			setupMocks: func(ms *service.IMerchantService, us *service.IUserService, rs *service.IRoleService, urs *service.IUserRoleService, rabbitMqExt *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "ERROR: failed to update user",
			userID:      "user-123",
			requestBody: []byte(`{"name": "John Doe", "email": "john.doe@example.com", "roleSlug": "admin", "status": "BLOCKED"}`),
			setupMocks: func(ms *service.IMerchantService, us *service.IUserService, rs *service.IRoleService, urs *service.IUserRoleService, rabbitMqExt *mockRabbitMq.RabbitMQExt) {
				us.On("FindUserByID", mock.Anything, "user-123").Return(&user.User{
					UUID:       "user-123",
					MerchantId: "merchant-123",
				}, nil)
				us.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to update user"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "ERROR: failed to get FindRoleBySlug",
			userID:      "user-123",
			requestBody: []byte(`{"name": "John Doe", "email": "john.doe@example.com", "roleSlug": "admin", "status": "BLOCKED"}`),
			setupMocks: func(ms *service.IMerchantService, us *service.IUserService, rs *service.IRoleService, urs *service.IUserRoleService, rabbitMqExt *mockRabbitMq.RabbitMQExt) {
				us.On("FindUserByID", mock.Anything, "user-123").Return(&user.User{
					UUID:       "user-123",
					MerchantId: "merchant-123",
				}, nil)
				us.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
				rs.On("FindRoleBySlug", mock.Anything, "admin").Return(nil, errors.New("failed to get FindRoleBySlug"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "ERROR: failed to get FindUserRoleByUserID",
			userID:      "user-123",
			requestBody: []byte(`{"name": "John Doe", "email": "john.doe@example.com", "roleSlug": "admin", "status": "BLOCKED"}`),
			setupMocks: func(ms *service.IMerchantService, us *service.IUserService, rs *service.IRoleService, urs *service.IUserRoleService, rabbitMqExt *mockRabbitMq.RabbitMQExt) {
				us.On("FindUserByID", mock.Anything, "user-123").Return(&user.User{
					UUID:       "user-123",
					MerchantId: "merchant-123",
					Role:       sql.NullString{String: "user", Valid: true},
				}, nil)
				us.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
				rs.On("FindRoleBySlug", mock.Anything, "admin").Return(&role.Role{
					UUID: "role-123",
					Slug: "admin",
				}, nil)
				urs.On("FindUserRoleByUserID", mock.Anything, "user-123").Return(nil, errors.New("failed to get FindUserRoleByUserID"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "ERROR: failed to update userRole UpdateByUserID",
			userID:      "user-123",
			requestBody: []byte(`{"name": "John Doe", "email": "john.doe@example.com", "roleSlug": "admin", "status": "BLOCKED"}`),
			setupMocks: func(ms *service.IMerchantService, us *service.IUserService, rs *service.IRoleService, urs *service.IUserRoleService, rabbitMqExt *mockRabbitMq.RabbitMQExt) {
				us.On("FindUserByID", mock.Anything, "user-123").Return(&user.User{
					UUID:       "user-123",
					MerchantId: "merchant-123",
					Role:       sql.NullString{String: "user", Valid: true},
				}, nil)
				us.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
				rs.On("FindRoleBySlug", mock.Anything, "admin").Return(&role.Role{
					UUID: "role-123",
					Slug: "admin",
				}, nil)
				urs.On("FindUserRoleByUserID", mock.Anything, "user-123").Return(&userRole.UserRole{
					UUID:   "user-role-123",
					UserID: "user-123",
					RoleID: "role-123",
				}, nil)
				urs.On("UpdateByUserID", mock.Anything, mock.AnythingOfType("*userRole.UserRole")).Return(errors.New("failed to update userRole UpdateByUserID"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockMerchantService := new(service.IMerchantService)
			mockUserService := new(service.IUserService)
			mockRoleService := new(service.IRoleService)
			mockUserRoleService := new(service.IUserRoleService)
			mockRabbit := mockRabbitMq.NewRabbitMQExt(t)
			mockValidator := validator.New()

			// Setup mocks
			tt.setupMocks(mockMerchantService, mockUserService, mockRoleService, mockUserRoleService, mockRabbit)

			// Create controller
			controller := &CRMUserController{
				config:      &config.Config{},
				secret:      &config.Secret{},
				merchantSvc: mockMerchantService,
				userSvc:     mockUserService,
				roleSvc:     mockRoleService,
				userRoleSvc: mockUserRoleService,
				validate:    mockValidator,
				rabbitMqExt: mockRabbit,
			}

			// Create request
			req, _ := http.NewRequest("PUT", "/users/"+tt.userID, bytes.NewReader(tt.requestBody))

			// Setup chi router context
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("user_id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the Update function
			controller.Update(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			// Assert expectations
			mockMerchantService.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
			mockRoleService.AssertExpectations(t)
			mockUserRoleService.AssertExpectations(t)
		})
	}
}
