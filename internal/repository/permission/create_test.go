package permissionRepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	now := time.Now()

	existedPermission := &permissionModel.Permission{
		UUID:      "uuid-uuid-uuid",
		Name:      "role-name",
		Slug:      "role-slug",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Define the test cases
	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		permission *permissionModel.Permission
		wantErr    bool
	}{
		{
			name:       "Valid User",
			permission: existedPermission,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*permissionModel.Permission"),
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
					mock.AnythingOfType("*permissionModel.Permission"),
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
					mock.AnythingOfType("*permissionModel.Permission"),
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
			err := repo.Create(context.Background(), tc.permission)

			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.Create() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
