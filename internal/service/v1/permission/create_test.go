package permissionService

import (
	"context"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	now := time.Now()

	createdPermission := &permissionModel.Permission{
		UUID:      "uuid-uuid-uuid",
		Name:      "admin",
		Slug:      "admin",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		name       string
		input      *permissionModel.Permission
		mocksSetup func(permissionRepo *repositoryMocks.IPermissionRepository)
		wantErr    bool
	}{
		{
			name:  "SUCCESS: successfully create permission",
			input: createdPermission,
			mocksSetup: func(permissionRepo *repositoryMocks.IPermissionRepository) {
				permissionRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*permissionModel.Permission"),
				).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			permissionRepo := repositoryMocks.NewIPermissionRepository(t)

			tc.mocksSetup(permissionRepo)

			trxSvc := New(permissionRepo, nil)

			ctx := context.Background()
			err := trxSvc.Create(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			permissionRepo.AssertExpectations(t)
		})
	}
}
