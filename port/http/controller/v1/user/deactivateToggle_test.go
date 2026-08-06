package user

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
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

func buildTestUserAndMerchant() (*merchantModel.Merchant, *userModel.User, *userModel.UserTokenClaims) {
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

	existedUser := &userModel.User{
		UUID:             uuid.NewString(),
		Email:            "test@gmail.com",
		Name:             "Testing",
		Status:           constant.UserStatusActive,
		Blocked:          sql.NullTime{Time: time.Time{}, Valid: false},
		MerchantId:       existedMerchant.UUID,
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

	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: existedMerchant.UUID,
	}

	return existedMerchant, existedUser, validUserClaims
}

type mockDeps struct {
	userSvc     *mockUser.IUserService
	roleSvc     *mockRole.IRoleService
	userRoleSvc *mockUserRole.IUserRoleService
	merchantSvc *mockMerchant.IMerchantService
	jwt         *mockJWT.IJwt
	rmq         *mockRabbitMq.RabbitMQExt
}

func newMockDeps(t *testing.T) mockDeps {
	return mockDeps{
		userSvc:     mockUser.NewIUserService(t),
		roleSvc:     mockRole.NewIRoleService(t),
		userRoleSvc: mockUserRole.NewIUserRoleService(t),
		merchantSvc: mockMerchant.NewIMerchantService(t),
		jwt:         mockJWT.NewIJwt(t),
		rmq:         mockRabbitMq.NewRabbitMQExt(t),
	}
}

func (d *mockDeps) controller() *UserController {
	cfg := &config.Config{ServiceName: "testing"}
	secret := &config.Secret{JWTSignatureKey: config.JWTSignatureKey{UserKey: "testing"}}
	return New(validator.New(), d.userSvc, d.roleSvc, d.userRoleSvc, d.merchantSvc, d.jwt, cfg, secret, d.rmq, nil)
}

func buildRequest(t *testing.T, userID string, userClaims *userModel.UserTokenClaims) *http.Request {
	t.Helper()
	chiRouterCtx := chi.NewRouteContext()
	chiRouterCtx.URLParams.Add("user_id", userID)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/{user_id}/activate", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
	if userClaims != nil {
		req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims))
	}
	return req
}

func TestUserController_ActivateUser(t *testing.T) {
	existedMerchant, existedUser, validUserClaims := buildTestUserAndMerchant()

	testCases := []struct {
		name           string
		userId         string
		mockSetup      func(d *mockDeps)
		expectedStatus int
		userClaims     *userModel.UserTokenClaims
	}{
		{
			name:   "SUCCESS",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(existedMerchant, nil).Once()
				d.userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(&userModel.User{
					UUID:          existedUser.UUID,
					Email:         existedUser.Email,
					Name:          existedUser.Name,
					Status:        constant.UserStatusInactive,
					MerchantId:    existedUser.MerchantId,
					MerchantName:  existedUser.MerchantName,
					Role:          existedUser.Role,
					DeactivatedAt: existedUser.DeactivatedAt,
					CreatedAt:     existedUser.CreatedAt,
					UpdatedAt:     existedUser.UpdatedAt,
				}, nil).Once()
				d.userSvc.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
				d.jwt.On("RemoveIterateTokenFromRedis", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything).Return(nil).Once()
			},
			expectedStatus: 200,
			userClaims:     validUserClaims,
		},
		{
			name:           "ERROR: User ID is required",
			mockSetup:      func(d *mockDeps) {},
			expectedStatus: http.StatusBadRequest,
			userClaims:     validUserClaims,
		},
		{
			name:   "ERROR: Failed to find merchant",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, errors.New("failed to find merchant")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:   "ERROR: Merchant not found",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, nil).Once()
			},
			expectedStatus: http.StatusNotFound,
			userClaims:     validUserClaims,
		},
		{
			name:   "ERROR: Service Error",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(existedMerchant, nil).Once()
				d.userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(&userModel.User{
					UUID:          existedUser.UUID,
					Email:         existedUser.Email,
					Name:          existedUser.Name,
					Status:        constant.UserStatusInactive,
					MerchantId:    existedUser.MerchantId,
					MerchantName:  existedUser.MerchantName,
					Role:          existedUser.Role,
					DeactivatedAt: existedUser.DeactivatedAt,
					CreatedAt:     existedUser.CreatedAt,
					UpdatedAt:     existedUser.UpdatedAt,
				}, nil).Once()
				d.userSvc.On("Update", mock.Anything, mock.Anything).Return(errors.New("service error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:           "ERROR: User not in Context",
			userId:         existedUser.UUID,
			mockSetup:      func(d *mockDeps) {},
			expectedStatus: http.StatusUnauthorized,
			userClaims:     nil,
		},
		{
			name:   "ERROR: User not found",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(existedMerchant, nil).Once()
				d.userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(nil, errors.New("user not found")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:   "ERROR: User not in the same merchant",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(existedMerchant, nil).Once()
				d.userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(&userModel.User{
					UUID:       uuid.NewString(),
					MerchantId: uuid.NewString(),
				}, nil).Once()
			},
			expectedStatus: http.StatusUnauthorized,
			userClaims:     validUserClaims,
		},
		{
			name:           "ERROR: Activate own account",
			userId:         validUserClaims.UUID,
			mockSetup:      func(d *mockDeps) {},
			expectedStatus: http.StatusBadRequest,
			userClaims:     validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			d := newMockDeps(t)
			tt.mockSetup(&d)
			mc := d.controller()

			req := buildRequest(t, tt.userId, tt.userClaims)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(mc.ActivateUser)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, rr.Code)
			d.userSvc.AssertExpectations(t)
		})
	}
}

func TestUserController_DeactivateUser(t *testing.T) {
	existedMerchant, existedUser, validUserClaims := buildTestUserAndMerchant()

	testCases := []struct {
		name           string
		userId         string
		mockSetup      func(d *mockDeps)
		expectedStatus int
		userClaims     *userModel.UserTokenClaims
	}{
		{
			name:   "SUCCESS",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(existedMerchant, nil)
				d.userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(existedUser, nil)
				d.userSvc.On("Update", mock.Anything, mock.Anything).Return(nil)
				d.jwt.On("RemoveIterateTokenFromRedis", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything).Return(nil)
			},
			expectedStatus: 200,
			userClaims:     validUserClaims,
		},
		{
			name:           "ERROR: User ID is required",
			mockSetup:      func(d *mockDeps) {},
			expectedStatus: http.StatusBadRequest,
			userClaims:     validUserClaims,
		},
		{
			name:   "ERROR: Failed to find merchant",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, errors.New("failed to find merchant"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:   "ERROR: Merchant not found",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			userClaims:     validUserClaims,
		},
		{
			name:   "ERROR: Service Error",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(existedMerchant, nil)
				d.userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(&userModel.User{
					UUID:          existedUser.UUID,
					Email:         existedUser.Email,
					Name:          existedUser.Name,
					Status:        constant.UserStatusActive,
					MerchantId:    existedUser.MerchantId,
					MerchantName:  existedUser.MerchantName,
					Role:          existedUser.Role,
					DeactivatedAt: existedUser.DeactivatedAt,
					CreatedAt:     existedUser.CreatedAt,
					UpdatedAt:     existedUser.UpdatedAt,
				}, nil)
				d.userSvc.On("Update", mock.Anything, mock.Anything).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:           "ERROR: User not in Context",
			userId:         existedUser.UUID,
			mockSetup:      func(d *mockDeps) {},
			expectedStatus: http.StatusUnauthorized,
			userClaims:     nil,
		},
		{
			name:   "ERROR: User not found",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(existedMerchant, nil)
				d.userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaims:     validUserClaims,
		},
		{
			name:   "ERROR: User not in the same merchant",
			userId: existedUser.UUID,
			mockSetup: func(d *mockDeps) {
				d.merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(existedMerchant, nil)
				d.userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(&userModel.User{
					UUID:       uuid.NewString(),
					MerchantId: uuid.NewString(),
				}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			userClaims:     validUserClaims,
		},
		{
			name:           "ERROR: Deactivate own account",
			userId:         validUserClaims.UUID,
			mockSetup:      func(d *mockDeps) {},
			expectedStatus: http.StatusBadRequest,
			userClaims:     validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			d := newMockDeps(t)
			tt.mockSetup(&d)
			mc := d.controller()

			req := buildRequest(t, tt.userId, tt.userClaims)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(mc.DeactivateUser)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, rr.Code)
			d.userSvc.AssertExpectations(t)
		})
	}
}
