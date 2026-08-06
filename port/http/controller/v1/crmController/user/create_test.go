package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/dictionary"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
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

func TestUserController_Create(t *testing.T) {
	expectedRole := &role.Role{
		UUID:      uuid.NewString(),
		Name:      "admin",
		Slug:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	existedMerchant := &merchantModel.Merchant{
		UUID:          uuid.NewString(),
		Name:          "merchant",
		Description:   "merchant description",
		Logo:          "merchant logo",
		MerchantEmail: "test@gmail.com",
		MerchantPhone: "08123456789",
		PICEmail:      "pic@gmail.com",
		PICPhone:      "08123456789",
		CreatedAt:     time.Now(),
	}

	payloadRequest := &userModel.CRMUserCreateRequest{
		Email:      "test@gmail.com",
		MerchantId: uuid.New().String(),
		Name:       "test",
		CreatedAt:  time.Now(),
		RoleSlug:   expectedRole.UUID,
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name             string
		requestBody      []byte
		expectedRole     *role.Role
		expectedMerchant *merchantModel.Merchant
		mockSetup        func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService)
		expectedStatus   int
		userClaims       *userModel.UserTokenClaims
	}{
		{
			name:             "SUCCESS",
			requestBody:      payloadRequestByte,
			expectedRole:     expectedRole,
			expectedMerchant: existedMerchant,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
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

				userSvc.
					On(
						"SendGeneratedInvitationURL",
						mock.Anything,
						mock.AnythingOfType("*user.SendGeneratedInvitationRequest")).
					Return(nil)
			},
			expectedStatus: 200,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"email": "12345abcde"}`),
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:         "ERROR: Failed to get role",
			requestBody:  payloadRequestByte,
			expectedRole: nil,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
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
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:         "ERROR: Role not found",
			requestBody:  payloadRequestByte,
			expectedRole: expectedRole,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
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
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:         "ERROR: Failed to create user role",
			requestBody:  payloadRequestByte,
			expectedRole: expectedRole,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
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
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:        "ERROR: Failed to find merchant",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(nil, errors.New("failed to find merchant"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:        "ERROR: Merchant not found",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:        "ERROR: Service Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
				userSvc.On("Create",
					mock.Anything,
					mock.Anything,
				).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:        "ERROR: FindUserByEmail Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:        "ERROR: Email already registered",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(&userModel.User{}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
		{
			name:        "ERROR: SendGeneratedInvitationURL Error",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByEmail",
						mock.Anything,
						mock.Anything).
					Return(nil, nil)
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

				userSvc.
					On(
						"SendGeneratedInvitationURL",
						mock.Anything,
						mock.AnythingOfType("*user.SendGeneratedInvitationRequest")).
					Return(errors.New("failed to send invitation"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims: &userModel.UserTokenClaims{
				UUID:       uuid.NewString(),
				Role:       "admin",
				MerchantId: uuid.NewString(),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName:      "testing",
				DictionaryConfig: config.DictionaryConfig{},
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

			tt.mockSetup(mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc)

			currentDir, err := os.Getwd()
			assert.NoError(t, err)

			projectName := "backend-portal"
			projectRoot, err := util.FindProjectRoot(currentDir, projectName)
			if err != nil {
				fmt.Printf("Error finding project root: %v\n", err)
				return
			}
			cfg.DictionaryConfig.Path = filepath.Join(projectRoot, "docs", "dictionary-i18n.json")

			// Dictionary
			dictionary.Dict, err = dictionary.New(cfg.DictionaryConfig)
			if err != nil {
				fmt.Printf("Unable to init dictionary, %v", err)
				panic(err)
			}

			mc := New(
				cfg, secret, mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc,
				WithJWT(jwtMock), WithRabbitMQClient(mockRmq), WithValidator(mockValidator),
			)

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBuffer(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaims))
			}

			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Create)
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
