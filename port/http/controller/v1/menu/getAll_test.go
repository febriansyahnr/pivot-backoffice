package menuController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestMenuController_GetAll(t *testing.T) {
	testCase := []struct {
		name           string
		queryParams    string
		expectedStatus int
		mockSetup      func(menuSvc *serviceMocks.IMenuService)
	}{
		{
			name:        "SUCCESS: Get all menus (without query param)",
			queryParams: "",
			mockSetup: func(menuSvc *serviceMocks.IMenuService) {
				menuSvc.On(
					"GetAll",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					false,
				).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "SUCCESS: Get all menus including Home (forRoleManagement=false)",
			queryParams: "?forRoleManagement=false",
			mockSetup: func(menuSvc *serviceMocks.IMenuService) {
				menuSvc.On(
					"GetAll",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					false,
				).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "SUCCESS: Get all menus excluding Home (forRoleManagement=true)",
			queryParams: "?forRoleManagement=true",
			mockSetup: func(menuSvc *serviceMocks.IMenuService) {
				menuSvc.On(
					"GetAll",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					true,
				).Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "ERROR: Get List service error",
			queryParams: "",
			mockSetup: func(menuSvc *serviceMocks.IMenuService) {
				menuSvc.On(
					"GetAll",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					false,
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			mockMenuSvc := serviceMocks.NewIMenuService(t)
			mockUserRoleSvc := serviceMocks.NewIUserRoleService(t)
			mockRoleSvc := serviceMocks.NewIRoleService(t)
			mockValidator := validator.New()

			tt.mockSetup(mockMenuSvc)

			cfg.AppConfig.PaginationPerPage = 20
			ctrl := New(cfg, mockValidator, mockMenuSvc, mockUserRoleSvc, mockRoleSvc)
			baseUrl := "/api/v1/menus" + tt.queryParams

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.GetAll)
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
