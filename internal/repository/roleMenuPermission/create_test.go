package roleMenuPermissionRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	roleMenuPermission := &roleMenuPermissionModel.RoleMenuPermission{
		RoleID:       "role-uuid-uuid-uuid",
		MenuID:       "menu-uuid-uuid-uuid",
		PermissionID: "permission-uuid-uuid-uuid",
	}

	// Define the test cases
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     *roleMenuPermissionModel.RoleMenuPermission
		wantErr   bool
	}{
		{
			name:  "Valid User",
			input: roleMenuPermission,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*roleMenuPermissionModel.RoleMenuPermission"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Insert to Database",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*roleMenuPermissionModel.RoleMenuPermission"),
				).Return(false, errors.New("insert error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*roleMenuPermissionModel.RoleMenuPermission"),
				).Return(false, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.Create(context.Background(), tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.Create() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
