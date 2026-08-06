package role

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	roleID := "4ce7611d-b97a-44b2-aa8d-9973b0680330"
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
		Role:       constant.RoleMaker,
	}

	testCases := []struct {
		name           string
		roleID         string
		mockSetup      func(roleSvc *mockRole.IRoleService)
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name:   "SUCCESS",
			roleID: roleID,
			mockSetup: func(roleSvc *mockRole.IRoleService) {
				roleSvc.On("Delete", constant.ValueCtxMockType(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name:   "ERROR:Service Delete error",
			roleID: roleID,
			mockSetup: func(roleSvc *mockRole.IRoleService) {
				roleSvc.On("Delete", constant.ValueCtxMockType(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(constant.ErrSomeErrorForUnitTest)
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:   "FAILED:Invalid Role ID",
			roleID: "invalid-role-id",
			mockSetup: func(*mockRole.IRoleService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:   "ERROR:Unauthorized",
			roleID: "invalid-role-id",
			mockSetup: func(*mockRole.IRoleService) {
				// Empty mock setup
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSvc := mockRole.NewIRoleService(t)

			tt.mockSetup(mockRoleSvc)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/roles/"+tt.roleID, nil)
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			router := chi.NewRouter()
			router.Delete(
				"/roles/{role_id}", New(validator.New(), mockRoleSvc, nil).Delete,
			)
			router.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockRoleSvc.AssertExpectations(t)
		})
	}
}
