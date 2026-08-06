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
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
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

func TestUserController_ListUsers(t *testing.T) {
	expectedUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "ganteng123",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  time.Now(),
	}

	expectedUsers := []*userModel.User{expectedUser}

	testCase := []struct {
		name           string
		mockSetup      func(userSvc *mockUser.IUserService)
		expectedStatus int
	}{
		{
			name: "SUCCESS",
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.On("ListUsers",
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int"),
				).Return(expectedUsers, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "ERROR: Service Error",
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.On("ListUsers",
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int"),
				).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCase {
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

			tt.mockSetup(mockUserSvc)

			mc := New(mockValidator, mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc, jwtMock, cfg, secret, mockRmq, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, "/users/list", nil)
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.ListUsers)
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

func TestFindByID(t *testing.T) {
	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
		Role:       constant.RoleMaker,
	}

	testCases := []struct {
		name           string
		mockSetup      func(userSvc *mockUser.IUserService)
		expectedStatus int
		userClaim      *userModel.UserTokenClaims
	}{
		{
			name: "SUCCESS",
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(&userModel.User{}, nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR",
			mockSetup: func(userSvc *mockUser.IUserService) {
				userSvc.On("FindUserByID", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockUserSvc := mockUser.NewIUserService(t)
			mockValidator := validator.New()

			tt.mockSetup(mockUserSvc)

			mc := New(mockValidator, mockUserSvc, nil, nil, nil, nil, nil, nil, nil, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, "/users/{user_id}", nil)
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Get(
				"/users/{user_id}", mc.FindByID,
			)
			router.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockUserSvc.AssertExpectations(t)
		})
	}
}
