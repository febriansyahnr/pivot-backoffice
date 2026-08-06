package role

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddDefaultRolePermissions(t *testing.T) {
	validPayload := &role.CRMUpdateDefaultRolePermissionsRequest{
		RoleSlug: "ADMIN",
		Menus: []role.RoleMenuRequest{
			{
				Slug:        "payment",
				Permissions: []string{"payment.link.create", "payment.link.read"},
			},
		},
	}

	invalidPayload := &role.CRMUpdateDefaultRolePermissionsRequest{
		RoleSlug: "INVALID_ROLE",
		Menus: []role.RoleMenuRequest{
			{
				Slug:        "payment",
				Permissions: []string{"payment.link.create"},
			},
		},
	}

	successResponse := &role.RoleMenuResponse{
		ID:   "role-uuid-123",
		Name: "Admin",
		Menus: []role.RoleMenuPermissionResponse{
			{
				Name:        "Accept Payments",
				Permissions: []string{"payment.link.create", "payment.link.read"},
			},
		},
	}

	testCases := []struct {
		name           string
		payload        interface{}
		mockSetup      func(roleSvc *mockService.IRoleService)
		expectedStatus int
	}{
		{
			name:    "SUCCESS: add permissions to default role",
			payload: validPayload,
			mockSetup: func(roleSvc *mockService.IRoleService) {
				roleSvc.On("AddDefaultRolePermissions", mock.Anything, mock.AnythingOfType("*role.CRMUpdateDefaultRolePermissionsRequest")).
					Return(successResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "ERROR: invalid JSON body",
			payload: "invalid json",
			mockSetup: func(roleSvc *mockService.IRoleService) {
				// No mock needed - should fail at JSON decode
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "ERROR: validation failed - invalid role slug",
			payload: invalidPayload,
			mockSetup: func(roleSvc *mockService.IRoleService) {
				// No mock needed - should fail at validation
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR: empty menus array",
			payload: &role.CRMUpdateDefaultRolePermissionsRequest{
				RoleSlug: "ADMIN",
				Menus:    []role.RoleMenuRequest{},
			},
			mockSetup: func(roleSvc *mockService.IRoleService) {
				// No mock needed - should fail at validation
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "ERROR: role not found",
			payload: validPayload,
			mockSetup: func(roleSvc *mockService.IRoleService) {
				roleSvc.On("AddDefaultRolePermissions", mock.Anything, mock.AnythingOfType("*role.CRMUpdateDefaultRolePermissionsRequest")).
					Return(nil, errors.New("role not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "ERROR: role is not a default role",
			payload: validPayload,
			mockSetup: func(roleSvc *mockService.IRoleService) {
				roleSvc.On("AddDefaultRolePermissions", mock.Anything, mock.AnythingOfType("*role.CRMUpdateDefaultRolePermissionsRequest")).
					Return(nil, errors.New("role is not a default role"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "ERROR: menu or permission not registered",
			payload: validPayload,
			mockSetup: func(roleSvc *mockService.IRoleService) {
				roleSvc.On("AddDefaultRolePermissions", mock.Anything, mock.AnythingOfType("*role.CRMUpdateDefaultRolePermissionsRequest")).
					Return(nil, errors.New("menu or permission not registered"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			roleSvc := mockService.NewIRoleService(t)
			validate := validator.New()
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mockSetup(roleSvc)

			// Create controller
			controller := New(roleSvc, validate, WithLogger(logger))

			// Create request
			var body []byte
			var err error
			if str, ok := tc.payload.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tc.payload)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/crm/roles/default-permissions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			controller.AddDefaultRolePermissions(w, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}

			roleSvc.AssertExpectations(t)
		})
	}
}
