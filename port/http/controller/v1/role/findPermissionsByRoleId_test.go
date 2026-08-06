package role

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRoleController_FindByRoleId(t *testing.T) {
	data := make([]*permissionModel.Permission, 0)
	data = append(data, &permissionModel.Permission{
		UUID:        uuid.NewString(),
		Slug:        "perm-test",
		Name:        "Perm Test",
		Description: "Perm testing",
		Group:       "perm-group",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	testCase := []struct {
		name           string
		roleId         string
		expectedStatus int
		mockSetup      func(roleSvc *mockRole.IRoleService, permMock *serviceMocks.IPermissionService)
	}{
		{
			name:   "SUCCESS: Get permissions by role id",
			roleId: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func(roleMock *mockRole.IRoleService, permMock *serviceMocks.IPermissionService) {
				permMock.On(
					"FindByRoleId",
					mock.AnythingOfType(constant.MockTypeValueContextReference), // Match the context type exactly
					mock.AnythingOfType("string"),
				).Return(data, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "FAILED: Got error 500 on get list caused by empty role id",
			mockSetup: func(roleMock *mockRole.IRoleService, permMock *serviceMocks.IPermissionService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "FAILED: Got error 500 on get list caused by service error",
			roleId: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func(roleMock *mockRole.IRoleService, permMock *serviceMocks.IPermissionService) {
				permMock.On(
					"FindByRoleId",
					mock.AnythingOfType(constant.MockTypeValueContextReference), // Match the context type exactly
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSvc := mockRole.NewIRoleService(t)
			mockPermissionSvc := serviceMocks.NewIPermissionService(t)
			mockValidator := validator.New()

			tt.mockSetup(mockRoleSvc, mockPermissionSvc)

			roleController := New(mockValidator, mockRoleSvc, mockPermissionSvc)
			baseUrl := "/api/v1/roles/:id/permissions"

			// Create the HTTP request for the test case
			chiRouterCtx := chi.NewRouteContext()
			chiRouterCtx.URLParams.Add("role_id", tt.roleId)
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(roleController.FindPermissionsByRoleId)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			mockRoleSvc.AssertExpectations(t)
		})
	}
}
