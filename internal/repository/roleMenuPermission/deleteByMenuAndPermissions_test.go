package roleMenuPermissionRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteByMenuAndPermissions(t *testing.T) {
	roleID := "role-uuid-123"
	menuID := "menu-uuid-1"
	permissionIDs := []string{"perm-uuid-1", "perm-uuid-2"}

	testCases := []struct {
		name          string
		roleID        string
		menuID        string
		permissionIDs []string
		mockSetup     func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr       bool
	}{
		{
			name:          "SUCCESS: delete permissions",
			roleID:        roleID,
			menuID:        menuID,
			permissionIDs: permissionIDs,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					roleID,
					menuID,
					"perm-uuid-1",
					"perm-uuid-2",
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:          "SUCCESS: no permissions to delete (empty array)",
			roleID:        roleID,
			menuID:        menuID,
			permissionIDs: []string{},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// No mock setup needed - should return early
			},
			wantErr: false,
		},
		{
			name:          "SUCCESS: delete single permission",
			roleID:        roleID,
			menuID:        menuID,
			permissionIDs: []string{"perm-uuid-1"},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					roleID,
					menuID,
					"perm-uuid-1",
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:          "ERROR: database error",
			roleID:        roleID,
			menuID:        menuID,
			permissionIDs: permissionIDs,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					roleID,
					menuID,
					"perm-uuid-1",
					"perm-uuid-2",
				).Return(false, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := mysqlMocks.NewIMySqlExt(t)
			logMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := New(mysqlMock, logMock)
			err := repo.DeleteByMenuAndPermissions(context.Background(), tc.roleID, tc.menuID, tc.permissionIDs)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}
