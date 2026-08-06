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

func TestUpdate(t *testing.T) {
	now := time.Now()

	existingPermission := &permissionModel.Permission{
		UUID:        "uuid-uuid-uuid",
		Name:        "role-name",
		Slug:        "role-slug",
		CreatedAt:   now,
		UpdatedAt:   now,
		Description: "",
		Group:       "",
	}

	// Define the test cases
	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		permission *permissionModel.Permission
		wantErr    bool
	}{
		{
			name:       "Valid",
			permission: existingPermission,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:       "ERROR: Failure Update to Database",
			permission: existingPermission,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, errors.New(" error"))
			},
			wantErr: true,
		},
		{
			name:       "ERROR: No Rows Affected",
			permission: existingPermission,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
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
			err := repo.Update(context.Background(), tc.permission)

			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.Update() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
