package roleMenuPermissionService

import (
	"context"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestCreate(t *testing.T) {
	roleMenuPermission := &roleMenuPermissionModel.RoleMenuPermission{
		RoleID:       "role-uuid-uuid-uuid",
		MenuID:       "menu-uuid-uuid-uuid",
		PermissionID: "permission-uuid-uuid-uuid",
	}

	testCases := []struct {
		name       string
		input      *roleMenuPermissionModel.RoleMenuPermission
		mocksSetup func(permissionRepo *repositoryMocks.IRoleMenuPermissionRepository)
		wantErr    bool
	}{
		{
			name:  "SUCCESS: successfully create permission",
			input: roleMenuPermission,
			mocksSetup: func(permissionRepo *repositoryMocks.IRoleMenuPermissionRepository) {
				permissionRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*roleMenuPermissionModel.RoleMenuPermission"),
				).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			roleMenuPermissionRepo := repositoryMocks.NewIRoleMenuPermissionRepository(t)

			tc.mocksSetup(roleMenuPermissionRepo)

			trxSvc := New(roleMenuPermissionRepo, nil)

			ctx := context.Background()
			err := trxSvc.Create(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			roleMenuPermissionRepo.AssertExpectations(t)
		})
	}
}
