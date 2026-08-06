package permissionService

import (
	"context"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestPermissionService_FindByRoleId(t *testing.T) {
	expectedPermission := &permissionModel.Permission{
		UUID:        uuid.NewString(),
		Slug:        "test-slug",
		Name:        "test-name",
		Description: "test-description",
		Group:       "test-group",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	list := make([]*permissionModel.Permission, 0)
	list = append(list, expectedPermission)

	testCases := []struct {
		name       string
		mocksSetup func(permissionRepo *repositoryMocks.IPermissionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS: find permission by role id",
			mocksSetup: func(permissionRepo *repositoryMocks.IPermissionRepository) {
				permissionRepo.On(
					"FindByRoleId",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(list, nil)
			},
			wantErr: false,
		},
		{
			name: "FAILED: got error service FindByRoleId",
			mocksSetup: func(permissionRepo *repositoryMocks.IPermissionRepository) {
				permissionRepo.On(
					"FindByRoleId",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: permissions not found",
			mocksSetup: func(permissionRepo *repositoryMocks.IPermissionRepository) {
				permissionRepo.On(
					"FindByRoleId",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			permissionRepo := repositoryMocks.NewIPermissionRepository(t)

			tc.mocksSetup(permissionRepo)

			trxSvc := New(permissionRepo, nil)

			ctx := context.Background()
			_, err := trxSvc.FindByRoleId(ctx, "role-id")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			permissionRepo.AssertExpectations(t)
		})
	}
}
