package role

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	roleID := "4ce7611d-b97a-44b2-aa8d-9973b0680330"
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
		Role:       constant.RoleMaker,
	}

	payloadRequest := &role.UpdateRoleRequest{
		Name: "Admin",
		Menus: []role.RoleMenuRequest{
			{
				Slug:        "dashboard-test",
				Permissions: []string{"dashboard-test.view"},
			},
		},
		MerchantID: validUserClaims.MerchantId,
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		roleID         string
		requestBody    []byte
		mockSetup      func(roleSvc *mockRole.IRoleService, permissionSvc *serviceMocks.IPermissionService)
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name:        "SUCCESS",
			roleID:      roleID,
			requestBody: payloadRequestByte,
			mockSetup: func(roleSvc *mockRole.IRoleService, permissionSvc *serviceMocks.IPermissionService) {
				roleSvc.On("UpdateRoleAndPermissions", mock.Anything, mock.Anything).Return(&role.RoleMenuResponse{}, nil)
			},
			expectedStatus: 200,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			roleID:      roleID,
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(userSvc *mockRole.IRoleService, permissionSvc *serviceMocks.IPermissionService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			roleID:      roleID,
			requestBody: []byte(`{"email": "12345abcde"}`),
			mockSetup: func(userSvc *mockRole.IRoleService, permissionSvc *serviceMocks.IPermissionService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Service Update error",
			roleID:      roleID,
			requestBody: payloadRequestByte,
			mockSetup: func(roleSvc *mockRole.IRoleService, permissionSvc *serviceMocks.IPermissionService) {
				roleSvc.On("UpdateRoleAndPermissions", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:        "FAILED: Invalid Role ID",
			roleID:      "invalid-role-id",
			requestBody: payloadRequestByte,
			mockSetup: func(*mockRole.IRoleService, *serviceMocks.IPermissionService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSvc := mockRole.NewIRoleService(t)
			mockPermissionSvc := serviceMocks.NewIPermissionService(t)
			mockValidator := validator.New()

			tt.mockSetup(mockRoleSvc, mockPermissionSvc)

			mc := New(mockValidator, mockRoleSvc, mockPermissionSvc)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPatch, "/roles/"+tt.roleID, bytes.NewBuffer(tt.requestBody))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Patch(
				"/roles/{role_id}", mc.Update,
			)
			router.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockRoleSvc.AssertExpectations(t)
		})
	}
}
