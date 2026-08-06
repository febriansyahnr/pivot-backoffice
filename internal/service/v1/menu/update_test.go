package menuService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name       string
		input      *role.Role
		mocksSetup func(menuRepo *repositoryMocks.IMenuRepository)
		wantErr    error
	}{
		{
			name: "when succeeded to update the menu, then should not return an error",
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"Update",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*menuModel.Menu"),
				).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			menuRepo := repositoryMocks.NewIMenuRepository(t)

			tc.mocksSetup(menuRepo)

			svc := New(menuRepo, nil)

			ctx := context.Background()
			err := svc.Update(ctx, &menuModel.Menu{})
			assert.Equal(t, tc.wantErr, err)

			menuRepo.AssertExpectations(t)
		})
	}
}

func TestIsShouldUpdate(t *testing.T) {
	testCases := []struct {
		name           string
		existingMenu   *menuModel.Menu
		newMenu        roleMenuPermissionModel.MenuPermissionFromFileRequest
		expectedResult bool
	}{
		{
			name: "when the name is different, then should return true",
			existingMenu: &menuModel.Menu{
				Name: "Old Name",
				Icon: "icon1",
			},
			newMenu: roleMenuPermissionModel.MenuPermissionFromFileRequest{
				Name: "New Name",
				Icon: "icon1",
			},
			expectedResult: true,
		},
		{
			name: "when the icon is different, then should return true",
			existingMenu: &menuModel.Menu{
				Name: "Menu Name",
				Icon: "icon1",
			},
			newMenu: roleMenuPermissionModel.MenuPermissionFromFileRequest{
				Name: "Menu Name",
				Icon: "icon2",
			},
			expectedResult: true,
		},
		{
			name: "when both name and icon are the same, then should return false",
			existingMenu: &menuModel.Menu{
				Name: "Menu Name",
				Icon: "icon1",
			},
			newMenu: roleMenuPermissionModel.MenuPermissionFromFileRequest{
				Name: "Menu Name",
				Icon: "icon1",
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := New(nil, nil)
			result := service.IsShouldUpdate(context.Background(), tc.existingMenu, tc.newMenu)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
