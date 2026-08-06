package role

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	claims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
		Role:       constant.RoleMaker,
	}
	mockRoleSvc := mockRole.NewIRoleService(t)
	mockValidator := validator.New()

	payload := &role.CreateRoleRequest{
		Name: "Tester",
		Menus: []role.RoleMenuRequest{
			{
				Slug:        "dashboard-test",
				Permissions: []string{"dashboard-test.view"},
			},
		},
		MerchantID: claims.MerchantId,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(roleSvc *mockRole.IRoleService)
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name:        "SUCCESS",
			requestBody: body,
			mockSetup: func(r *mockRole.IRoleService) {
				r.On(
					"CreateRoleAndPermissions", constant.ValueCtxMockType(), mock.AnythingOfType("*role.CreateRoleRequest"),
				).Once().Return(&role.RoleMenuResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      claims,
		},
		{
			name:        "ERROR:Bad Request/Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(*mockRole.IRoleService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      claims,
		},
		{
			name:        "ERROR:Bad Request/Failed Validation",
			requestBody: []byte(`{"email": "12345abcde"}`),
			mockSetup: func(*mockRole.IRoleService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      claims,
		},
		{
			name:        "ERROR:Service Error",
			requestBody: body,
			mockSetup: func(r *mockRole.IRoleService) {
				r.On(
					"CreateRoleAndPermissions", constant.ValueCtxMockType(), mock.AnythingOfType("*role.CreateRoleRequest"),
				).Once().Return(nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("invalid data")))
			},
			expectedStatus: http.StatusUnprocessableEntity,
			userClaim:      claims,
		},
		{
			name: "ERROR:User not in Context",
			mockSetup: func(*mockRole.IRoleService) {
				// Empty mock setup
			},
			userClaim:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			tt.mockSetup(mockRoleSvc)

			mc := New(mockValidator, mockRoleSvc, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/roles/create", bytes.NewBuffer(tt.requestBody))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Create)
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
