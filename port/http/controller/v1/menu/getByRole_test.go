package menuController

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	mockUserRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/userRole"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestMenuController_GetByRole(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	testCase := []struct {
		name           string
		expectedStatus int
		mockSetup      func(menuSvc *serviceMocks.IMenuService, userRoleSvc *mockUserRole.IUserRoleService)
		userClaim      *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Get List",
			mockSetup: func(menuSvc *serviceMocks.IMenuService, userRoleSvc *mockUserRole.IUserRoleService) {
				userRoleSvc.On(
					"FindUserRoleByUserID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&userRole.UserRole{RoleID: uuid.NewString()}, nil)

				menuSvc.On(
					"GetByRole",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("bool"),
				).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR: FindUserRoleByUserID service",
			mockSetup: func(menuSvc *serviceMocks.IMenuService, userRoleSvc *mockUserRole.IUserRoleService) {
				userRoleSvc.On(
					"FindUserRoleByUserID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR: GetByRole service",
			mockSetup: func(menuSvc *serviceMocks.IMenuService, userRoleSvc *mockUserRole.IUserRoleService) {
				userRoleSvc.On(
					"FindUserRoleByUserID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&userRole.UserRole{RoleID: uuid.NewString()}, nil)

				menuSvc.On(
					"GetByRole",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("bool"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR: User not in Context",
			mockSetup: func(menuSvc *serviceMocks.IMenuService, userRoleSvc *mockUserRole.IUserRoleService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusUnauthorized,
			userClaim:      nil,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			mockMenuSvc := serviceMocks.NewIMenuService(t)
			mockUserRoleSvc := mockUserRole.NewIUserRoleService(t)
			mockRoleSvc := serviceMocks.NewIRoleService(t)
			mockValidator := validator.New()

			tt.mockSetup(mockMenuSvc, mockUserRoleSvc)

			cfg.AppConfig.PaginationPerPage = 20
			ctrl := New(cfg, mockValidator, mockMenuSvc, mockUserRoleSvc, mockRoleSvc)
			baseUrl := "/api/v1/menus/role"

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.GetByRole)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockMenuSvc.AssertExpectations(t)
			mockUserRoleSvc.AssertExpectations(t)
		})
	}
}
