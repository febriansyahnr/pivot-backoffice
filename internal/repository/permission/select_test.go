package permissionRepository

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPermissionRepository_FindBySlug(t *testing.T) {
	now := time.Now()

	existedPermission := &permissionModel.Permission{
		UUID:      "uuid-uuid-uuid",
		Name:      "role-name",
		Slug:      "role-slug",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		name        string
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		input       string
		expected    *permissionModel.Permission
		expectedErr string
		wantErr     bool
	}{
		{
			name: "SUCCESS: Find Role By ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*permissionModel.Permission"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					permissionPtr := args.Get(1).(*permissionModel.Permission)
					*permissionPtr = *existedPermission
				})
			},
			input:       existedPermission.Slug,
			expected:    existedPermission,
			expectedErr: "",
			wantErr:     false,
		},
		{
			name: "ERROR: Permission Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*permissionModel.Permission"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			input:    existedPermission.Slug,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database FindBySlug Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*permissionModel.Permission"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			input:       existedPermission.Slug,
			expected:    nil,
			expectedErr: constant.ErrSomeErrorForUnitTest.Error(),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mysqlMock)

			repo := New(mysqlMock, loggerMock)

			ctx := context.Background()

			roleRes, err := repo.FindBySlug(ctx, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, roleRes)
				require.True(t, strings.Contains(err.Error(), tc.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, roleRes)
			}
		})
	}
}

func TestPermissionRepository_FindByRoleId(t *testing.T) {
	now := time.Now()
	roleId := uuid.NewString()

	existedPermission := &permissionModel.Permission{
		UUID:      "uuid-uuid-uuid",
		Name:      "role-name",
		Slug:      "role-slug",
		CreatedAt: now,
		UpdatedAt: now,
	}

	list := make([]*permissionModel.Permission, 0)
	list = append(list, existedPermission)

	testCases := []struct {
		name      string
		expected  []*permissionModel.Permission
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name:     "SUCCESS: Find permission by role id",
			expected: list,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					permissionPtr := args.Get(1).(*[]*permissionModel.Permission)
					*permissionPtr = list
				})
			},
			wantErr: false,
		},
		{
			name:     "ERROR: Permission not found",
			expected: nil,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
		{
			name:     "ERROR: Database FindByRoleId Error",
			expected: nil,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := New(mysqlMock, loggerMock)

			ctx := context.Background()

			res, err := repo.FindByRoleId(ctx, roleId)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}
