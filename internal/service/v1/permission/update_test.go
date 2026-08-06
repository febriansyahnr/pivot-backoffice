package permissionService

import (
	"context"
	"testing"
	"time"

	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	now := time.Now()

	UpdatedPermission := &permissionModel.Permission{
		UUID:        "uuid-uuid-uuid",
		Name:        "admin",
		Slug:        "admin",
		Description: "admin",
		Group:       "admin",
		UpdatedAt:   now,
	}

	testCases := []struct {
		name       string
		input      *permissionModel.Permission
		mocksSetup func(permissionRepo *repositoryMocks.IPermissionRepository)
		wantErr    bool
	}{
		{
			name:  "SUCCESS: successfully Update permission",
			input: UpdatedPermission,
			mocksSetup: func(permissionRepo *repositoryMocks.IPermissionRepository) {
				permissionRepo.On(
					"Update",
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
			err := trxSvc.Update(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			permissionRepo.AssertExpectations(t)
		})
	}
}
