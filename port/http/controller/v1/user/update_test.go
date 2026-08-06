package user

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"
	mockUserRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/userRole"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserController_Update(t *testing.T) {
	merchantId := uuid.NewString()
	expectedRole := &role.Role{
		UUID:      uuid.NewString(),
		Name:      "admin",
		Slug:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	existedUser := &userModel.User{
		UUID:             uuid.NewString(),
		Email:            "test@gmail.com",
		Name:             "Testing",
		Blocked:          sql.NullTime{Time: time.Time{}, Valid: false},
		MerchantId:       merchantId,
		MerchantName:     "merchant",
		RefreshToken:     sql.NullString{String: "refresh-token", Valid: false},
		IsChangePassword: 0,
		Role:             sql.NullString{String: "admin", Valid: true},
		LastLoginAt:      commonModel.CustomNullTime{NullTime: sql.NullTime{Time: time.Now(), Valid: false}},
		DeactivatedAt:    sql.NullTime{Time: time.Time{}, Valid: false},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        sql.NullTime{Time: time.Time{}, Valid: false},
		Password:         "password",
		PinHash:          sql.NullString{String: "pin-hash", Valid: false},
	}

	existedUserNonAdmin := &userModel.User{
		UUID:             uuid.NewString(),
		Email:            "test@gmail.com",
		Name:             "Testing",
		Blocked:          sql.NullTime{Time: time.Time{}, Valid: false},
		MerchantId:       merchantId,
		MerchantName:     "merchant",
		RefreshToken:     sql.NullString{String: "refresh-token", Valid: false},
		IsChangePassword: 0,
		Role:             sql.NullString{String: "approver", Valid: true},
		LastLoginAt:      commonModel.CustomNullTime{NullTime: sql.NullTime{Time: time.Now(), Valid: false}},
		DeactivatedAt:    sql.NullTime{Time: time.Time{}, Valid: false},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        sql.NullTime{Time: time.Time{}, Valid: false},
		Password:         "password",
		PinHash:          sql.NullString{String: "pin-hash", Valid: false},
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

	existedUserRole := &userRole.UserRole{
		UUID:      uuid.NewString(),
		UserID:    existedUser.UUID,
		RoleID:    expectedRole.UUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
	}

	payloadRequest := &userModel.UserCreateRequest{
		Email:     "test@gmail.com",
		Name:      "test",
		CreatedAt: time.Now(),
		RoleSlug:  expectedRole.Slug,
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	payloadRequestNonAdmin := &userModel.UserCreateRequest{
		Email:     "test@gmail.com",
		Name:      "test",
		CreatedAt: time.Now(),
		RoleSlug:  "approver",
	}
	payloadRequestByteNonAdmin, err := json.Marshal(payloadRequestNonAdmin)
	assert.NoError(t, err)

	validUserClaims := &userModel.UserTokenClaims{
		UUID:       existedUser.UUID,
		MerchantId: existedUser.MerchantId,
	}

	testCases := []struct {
		name             string
		userId           string
		requestBody      []byte
		expectedRole     *role.Role
		expectedMerchant *merchantModel.Merchant
		mockSetup        func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt)
		expectedStatus   int
		userClaims       *userModel.UserTokenClaims
	}{
		{
			name:             "SUCCESS",
			userId:           existedUserNonAdmin.UUID,
			requestBody:      payloadRequestByte,
			expectedRole:     expectedRole,
			expectedMerchant: existedMerchant,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedUserNonAdmin, nil)
				userSvc.
					On(
						"Update",
						mock.Anything,
						mock.Anything).
					Once().Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Once().Return(expectedRole, nil)
				userRole.
					On(
						"FindUserRoleByUserID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedUserRole, nil)
				userRole.
					On(
						"UpdateByUserID",
						mock.Anything,
						mock.AnythingOfType("*userRole.UserRole")).
					Once().Return(nil)
				jwt.On("TerminateTokenOfUserRoleChanged", mock.Anything, "test@gmail.com").Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: 200,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: User ID is required",
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			userId:      existedUser.UUID,
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			userId:      existedUser.UUID,
			requestBody: []byte(`{"email": "12345abcde"}`),
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaims:     validUserClaims,
		},
		{
			name:         "ERROR: Failed to get role",
			userId:       existedUser.UUID,
			requestBody:  payloadRequestByte,
			expectedRole: nil,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedUser, nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Once().Return(nil, errors.New("failed to get role"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:         "ERROR: Role not found",
			userId:       existedUser.UUID,
			requestBody:  payloadRequestByte,
			expectedRole: expectedRole,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedUser, nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Once().Return(nil, errors.New("role not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:         "ERROR: Failed to update user role",
			userId:       existedUserNonAdmin.UUID,
			requestBody:  payloadRequestByte,
			expectedRole: expectedRole,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedUserNonAdmin, nil)
				userSvc.
					On(
						"Update",
						mock.Anything,
						mock.Anything).
					Once().Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Once().Return(expectedRole, nil)
				userRole.
					On(
						"FindUserRoleByUserID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedUserRole, nil)
				userRole.
					On(
						"UpdateByUserID",
						mock.Anything,
						mock.AnythingOfType("*userRole.UserRole")).
					Once().Return(errors.New("failed to update user role"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: Failed to find merchant",
			userId:      existedUser.UUID,
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(nil, errors.New("failed to find merchant"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: Merchant not found",
			userId:      existedUser.UUID,
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: Service Error",
			userId:      existedUser.UUID,
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedUser, nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Once().Return(expectedRole, nil)
				userSvc.
					On(
						"Update",
						mock.Anything,
						mock.Anything).
					Once().Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: User not in Context",
			userId:      existedUser.UUID,
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
			},
			expectedStatus: http.StatusUnauthorized,
			userClaims:     nil,
		},
		{
			name:        "ERROR: User not found",
			userId:      existedUser.UUID,
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Once().Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:        "ERROR: User Role not found",
			userId:      existedUserNonAdmin.UUID,
			requestBody: payloadRequestByte,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedUserNonAdmin, nil)
				userSvc.
					On(
						"Update",
						mock.Anything,
						mock.Anything).
					Once().Return(nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Once().Return(expectedRole, nil)
				userRole.
					On(
						"FindUserRoleByUserID",
						mock.Anything,
						mock.Anything).
					Once().Return(nil, errors.New("user role not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:             "ERROR: user is not in the same merchant",
			userId:           existedUser.UUID,
			requestBody:      payloadRequestByte,
			expectedRole:     expectedRole,
			expectedMerchant: existedMerchant,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Return(&userModel.User{
						UUID:             uuid.NewString(),
						Email:            "test@gmail.com",
						Name:             "Testing",
						Blocked:          sql.NullTime{Time: time.Time{}, Valid: false},
						MerchantId:       uuid.NewString(),
						MerchantName:     "merchant",
						RefreshToken:     sql.NullString{String: "refresh-token", Valid: false},
						IsChangePassword: 0,
						Role:             sql.NullString{String: "admin", Valid: true},
						LastLoginAt:      commonModel.CustomNullTime{NullTime: sql.NullTime{Time: time.Now(), Valid: false}},
						DeactivatedAt:    sql.NullTime{Time: time.Time{}, Valid: false},
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
						DeletedAt:        sql.NullTime{Time: time.Time{}, Valid: false},
						Password:         "password",
						PinHash:          sql.NullString{String: "pin-hash", Valid: false},
					}, nil)
			},
			expectedStatus: 401,
			userClaims:     validUserClaims,
		},
		{
			name:             "ERROR: user is admin and trying to change own role",
			userId:           existedUser.UUID,
			requestBody:      payloadRequestByteNonAdmin,
			expectedRole:     expectedRole,
			expectedMerchant: existedMerchant,
			mockSetup: func(userSvc *mockUser.IUserService, roleSvc *mockRole.IRoleService, userRole *mockUserRole.IUserRoleService, merchantSvc *mockMerchant.IMerchantService, jwt *mockJWT.IJwt) {
				merchantSvc.
					On(
						"FindMerchantByID",
						mock.Anything,
						mock.Anything).
					Once().Return(existedMerchant, nil)
				userSvc.
					On(
						"FindUserByID",
						mock.Anything,
						mock.Anything).
					Return(existedUser, nil)
				roleSvc.
					On(
						"FindRoleBySlug",
						mock.Anything,
						mock.Anything).
					Once().Return(expectedRole, nil)
			},
			expectedStatus: 403,
			userClaims:     validUserClaims,
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
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tt.mockSetup(mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc, jwtMock)

			mc := New(mockValidator, mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc, jwtMock, cfg, secret, mockRmq, loggerMock)

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			chiRouterCtx.URLParams.Add("user_id", tt.userId)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users/{id}", bytes.NewBuffer(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaims))
			}

			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Update)
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
