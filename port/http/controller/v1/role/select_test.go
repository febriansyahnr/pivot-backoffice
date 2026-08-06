package role

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	roleModel "github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoleController_GetList(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
		Role:       constant.RoleMaker,
	}

	data := make([]roleModel.Role, 0)
	data = append(data, roleModel.Role{
		UUID:      "",
		Name:      "",
		Slug:      "",
		Type:      "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

	testCase := []struct {
		name            string
		expectedStatus  int
		mockSetup       func(roleSvc *mockRole.IRoleService)
		funcQueryParams func() *url.Values
		userClaim       *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mockService *mockRole.IRoleService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference), // Match the context type exactly
					mock.AnythingOfType("*role.GetRoleFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaim: validUserClaims,
		},
		{
			name: "SUCCESS: Get role that has page values",
			mockSetup: func(mockService *mockRole.IRoleService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference), // Match the context type exactly
					mock.AnythingOfType("*role.GetRoleFilterRequest"),
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
			userClaim: validUserClaims,
		},
		{
			name: "FAILED: Got error 500 on get list caused by service error",
			mockSetup: func(mockService *mockRole.IRoleService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference), // Match the context type exactly
					mock.AnythingOfType("*role.GetRoleFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("some-error"))
			},
			expectedStatus: http.StatusInternalServerError,
			funcQueryParams: func() *url.Values {
				return nil
			},
			userClaim: validUserClaims,
		},
		{
			name:           "FAILED: Got error 400 on get list caused by invalid page",
			mockSetup:      func(mockService *mockRole.IRoleService) {},
			expectedStatus: http.StatusBadRequest,
			funcQueryParams: func() *url.Values {
				queryParams := url.Values{}
				queryParams.Add("page", "invalid format")
				return &queryParams
			},
			userClaim: validUserClaims,
		},
		{
			name:           "ERROR: User not in Context",
			mockSetup:      func(_ *mockRole.IRoleService) {},
			userClaim:      nil,
			expectedStatus: http.StatusUnauthorized,
			funcQueryParams: func() *url.Values {
				return nil
			},
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			mockRoleSvc := mockRole.NewIRoleService(t)
			mockPermissionSvc := serviceMocks.NewIPermissionService(t)
			mockValidator := validator.New()

			tt.mockSetup(mockRoleSvc)

			cfg.AppConfig.PaginationPerPage = 20
			roleController := New(mockValidator, mockRoleSvc, mockPermissionSvc)
			baseUrl := "/api/v1/roles"

			if tt.funcQueryParams() != nil {
				baseUrl += "?" + tt.funcQueryParams().Encode()
			}

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(roleController.GetList)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockRoleSvc.AssertExpectations(t)
		})
	}
}
