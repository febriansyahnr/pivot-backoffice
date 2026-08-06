package user

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"
	mockUserRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/userRole"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserController_ListByMerchantID(t *testing.T) {
	data := make([]userModel.User, 0)
	data = append(data, userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "ganteng123",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  time.Now(),
	})

	expectedResponse := &commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	testCase := []struct {
		name            string
		expectedStatus  int
		mockSetup       func(userSvc *mockUser.IUserService)
		funcQueryParams func() *url.Values
		userClaims      *userModel.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mockService *mockUser.IUserService) {
				mockService.On(
					"ListUsersByMerchantID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.ListUsersByMerchantIDRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List with created_at filter",
			mockSetup: func(mockService *mockUser.IUserService) {
				mockService.On(
					"ListUsersByMerchantID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.ListUsersByMerchantIDRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", util.TimeNow.Add(-24*time.Hour).Format(util.UTCLayout))
				queryParams.Add("endCreatedAt", util.TimeNow.Format(util.UTCLayout))
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "SUCCESS: Get List that has page values",
			mockSetup: func(mockService *mockUser.IUserService) {
				mockService.On(
					"ListUsersByMerchantID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.ListUsersByMerchantIDRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "2")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 500 on get list caused by service error",
			mockSetup: func(mockService *mockUser.IUserService) {
				mockService.On(
					"ListUsersByMerchantID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.ListUsersByMerchantIDRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("some-error"))
			},
			expectedStatus: http.StatusInternalServerError,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid startCreatedAt",
			mockSetup: func(mockService *mockUser.IUserService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("startCreatedAt", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid endCreatedAt",
			mockSetup: func(mockService *mockUser.IUserService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("endCreatedAt", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid page",
			mockSetup: func(mockService *mockUser.IUserService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "invalid format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "FAILED: Got error 400 on get list caused by invalid perPage",
			mockSetup: func(mockService *mockUser.IUserService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("perPage", "invalid perPage format")
				return &queryParams
			},
			userClaims: validUserClaims,
		},
		{
			name: "ERROR: User not in Context",
			mockSetup: func(mockService *mockUser.IUserService) {
				// Empty mock setup
			},
			userClaims: nil,
			funcQueryParams: func() *url.Values {
				return nil
			},
			expectedStatus: http.StatusUnauthorized,
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

			cfg.AppConfig.PaginationPerPage = 20
			userController := New(mockValidator, mockUserSvc, mockRoleSvc, userRoleMock, mockMerchantSvc, jwtMock, cfg, secret, mockRmq, nil)
			baseUrl := "/api/v1/users"

			// Append query parameters to the URL
			if tt.funcQueryParams() != nil {
				baseUrl += "?" + tt.funcQueryParams().Encode()
			}

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaims))
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(userController.ListByMerchantID)
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
