package permissionService

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestPermissionService_FindBySlug(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(permissionRepo *repositoryMocks.IPermissionRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS: find permission by slug",
			mocksSetup: func(permissionRepo *repositoryMocks.IPermissionRepository) {
				permissionRepo.On(
					"FindBySlug",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&permissionModel.Permission{}, nil)
			},
			wantErr: false,
		},
		{
			name: "FAILED: got error service FindBySlug",
			mocksSetup: func(permissionRepo *repositoryMocks.IPermissionRepository) {
				permissionRepo.On(
					"FindBySlug",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: empty permission",
			mocksSetup: func(permissionRepo *repositoryMocks.IPermissionRepository) {
				permissionRepo.On(
					"FindBySlug",
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
			_, err := trxSvc.FindBySlug(ctx, "slug")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			permissionRepo.AssertExpectations(t)
		})
	}
}
