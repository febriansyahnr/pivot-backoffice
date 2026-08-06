package menuController

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockUserRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/userRole"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMenuController_GetByRoleId(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	testCase := []struct {
		name           string
		roleId         string
		expectedStatus int
		mockSetup      func(menuSvc *serviceMocks.IMenuService, roleSvc *serviceMocks.IRoleService)
		userClaim      *user.UserTokenClaims
	}{
		{
			name:   "SUCCESS: Get by role_id",
			roleId: uuid.NewString(),
			mockSetup: func(menuSvc *serviceMocks.IMenuService, roleSvc *serviceMocks.IRoleService) {
				roleSvc.On(
					"FindRoleById",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&role.Role{
					UUID:       uuid.NewString(),
					MerchantID: sql.NullString{String: validUserClaims.MerchantId, Valid: true},
				}, nil)

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
			name:   "ERROR: Failed to validate role_id",
			roleId: uuid.NewString(),
			mockSetup: func(menuSvc *serviceMocks.IMenuService, roleSvc *serviceMocks.IRoleService) {
				roleSvc.On(
					"FindRoleById",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:   "ERROR: Role Merchant ID is not the same with User Merchant ID",
			roleId: uuid.NewString(),
			mockSetup: func(menuSvc *serviceMocks.IMenuService, roleSvc *serviceMocks.IRoleService) {
				roleSvc.On(
					"FindRoleById",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&role.Role{
					UUID:       uuid.NewString(),
					MerchantID: sql.NullString{String: uuid.NewString(), Valid: true},
					Type:       constant.RoleTypeCustom,
				}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			userClaim:      validUserClaims,
		},
		{
			name:   "ERROR: User not in Context",
			roleId: uuid.NewString(),
			mockSetup: func(menuSvc *serviceMocks.IMenuService, roleSvc *serviceMocks.IRoleService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusUnauthorized,
			userClaim:      nil,
		},
		{
			name:   "ERROR: Failed to get menu by role",
			roleId: uuid.NewString(),
			mockSetup: func(menuSvc *serviceMocks.IMenuService, roleSvc *serviceMocks.IRoleService) {
				roleSvc.On(
					"FindRoleById",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&role.Role{
					UUID:       uuid.NewString(),
					MerchantID: sql.NullString{String: validUserClaims.MerchantId, Valid: true},
				}, nil)

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
			name: "ERROR: role_id is required",
			mockSetup: func(menuSvc *serviceMocks.IMenuService, roleSvc *serviceMocks.IRoleService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
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

			tt.mockSetup(mockMenuSvc, mockRoleSvc)

			cfg.AppConfig.PaginationPerPage = 20
			ctrl := New(cfg, mockValidator, mockMenuSvc, mockUserRoleSvc, mockRoleSvc)
			baseUrl := "/api/v1/menus/role/{role_id}"

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			chiRouterCtx.URLParams.Add("role_id", tt.roleId)

			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.GetByRoleId)
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
