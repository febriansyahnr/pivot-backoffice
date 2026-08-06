package roleMenuPermissionRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByRoleID(t *testing.T) {
	roleID := "role-uuid-123"

	expectedPermissions := []*roleMenuPermissionModel.RoleMenuPermission{
		{
			RoleID:       roleID,
			MenuID:       "menu-uuid-1",
			PermissionID: "perm-uuid-1",
		},
		{
			RoleID:       roleID,
			MenuID:       "menu-uuid-2",
			PermissionID: "perm-uuid-2",
		},
	}

	testCases := []struct {
		name      string
		roleID    string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
		expected  int
	}{
		{
			name:   "SUCCESS: get permissions by role ID",
			roleID: roleID,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*roleMenuPermissionModel.RoleMenuPermission"),
					mock.AnythingOfType("string"),
					roleID,
				).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*[]*roleMenuPermissionModel.RoleMenuPermission)
					*arg = expectedPermissions
				}).Return(nil)
			},
			wantErr:  false,
			expected: 2,
		},
		{
			name:   "SUCCESS: no permissions found",
			roleID: "non-existent-role",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*roleMenuPermissionModel.RoleMenuPermission"),
					mock.AnythingOfType("string"),
					"non-existent-role",
				).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*[]*roleMenuPermissionModel.RoleMenuPermission)
					*arg = []*roleMenuPermissionModel.RoleMenuPermission{}
				}).Return(nil)
			},
			wantErr:  false,
			expected: 0,
		},
		{
			name:   "ERROR: database error",
			roleID: roleID,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*roleMenuPermissionModel.RoleMenuPermission"),
					mock.AnythingOfType("string"),
					roleID,
				).Return(errors.New("database error"))
			},
			wantErr:  true,
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := mysqlMocks.NewIMySqlExt(t)
			logMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := New(mysqlMock, logMock)
			result, err := repo.GetByRoleID(context.Background(), tc.roleID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expected, len(result))
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}
