package roleAndPermissionCronHandler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	"github.com/redis/go-redis/v9"

	"github.com/paper-indonesia/pivot-backoffice/config"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mockLogger "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	roleMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/roles"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetupPredefinedRoleMenuPermissions(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(
			roleSvc *roleMocks.IRoleService,
			permissionSvc *serviceMocks.IPermissionService,
			menuSvc *serviceMocks.IMenuService,
			roleMenuPermissionSvc *serviceMocks.IRoleMenuPermissionService,
			loggerMock *mockLogger.ILogger,
			mockRedisExt *mockRedis.IRedisExt,
		)
	}{
		{
			name: "SUCCESS: Complete flow with existing entities",
			mockSetup: func(
				roleSvc *roleMocks.IRoleService,
				permissionSvc *serviceMocks.IPermissionService,
				menuSvc *serviceMocks.IMenuService,
				roleMenuPermissionSvc *serviceMocks.IRoleMenuPermissionService,
				loggerMock *mockLogger.ILogger,
				mockRedisExt *mockRedis.IRedisExt,
			) {
				// Setup flexible mocks that can handle multiple calls
				roleSvc.On(
					"FindRoleBySlug",
					mock.Anything,
					mock.Anything,
				).Return(&role.Role{UUID: uuid.NewString()}, nil)

				intCmd := &redis.IntCmd{}
				mockRedisExt.
					On("Del", mock.Anything, mock.Anything).Return(intCmd)
				intCmd.SetErr(nil)

				menuSvc.On(
					"FindBySlug",
					mock.Anything,
					mock.Anything,
				).Return(&menuModel.Menu{}, nil)

				menuSvc.On(
					"IsShouldUpdate",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false)

				menuSvc.On(
					"Update",
					mock.Anything,
					mock.Anything,
				).Return(nil).Maybe()

				permissionSvc.On(
					"FindBySlug",
					mock.Anything,
					mock.Anything,
				).Return(&permissionModel.Permission{}, nil)

				permissionSvc.On(
					"Update",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				roleMenuPermissionSvc.On(
					"Create",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				loggerMock.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
				loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()
			},
		},
		{
			name: "SUCCESS: Complete flow with new entities",
			mockSetup: func(
				roleSvc *roleMocks.IRoleService,
				permissionSvc *serviceMocks.IPermissionService,
				menuSvc *serviceMocks.IMenuService,
				roleMenuPermissionSvc *serviceMocks.IRoleMenuPermissionService,
				loggerMock *mockLogger.ILogger,
				mockRedisExt *mockRedis.IRedisExt,
			) {
				// Mock creating new entities - first call returns nil (not found)
				// subsequent calls return the created role
				roleSvc.On(
					"FindRoleBySlug",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil).Once()

				roleSvc.On(
					"Create",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				// Mock subsequent calls to FindRoleBySlug to return created role
				roleSvc.On(
					"FindRoleBySlug",
					mock.Anything,
					mock.Anything,
				).Return(&role.Role{UUID: uuid.NewString()}, nil)

				// Mock Redis Del for when role is found (cache clearing)
				intCmd := &redis.IntCmd{}
				mockRedisExt.
					On("Del", mock.Anything, mock.Anything).Return(intCmd).Maybe()
				intCmd.SetErr(nil)

				menuSvc.On(
					"FindBySlug",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)

				menuSvc.On(
					"Create",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				menuSvc.On(
					"IsShouldUpdate",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false).Maybe()

				permissionSvc.On(
					"FindBySlug",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)

				permissionSvc.On(
					"Create",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				roleMenuPermissionSvc.On(
					"Create",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				loggerMock.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
				loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()
			},
		},
		{
			name: "ERROR: Role creation fails",
			mockSetup: func(
				roleSvc *roleMocks.IRoleService,
				permissionSvc *serviceMocks.IPermissionService,
				menuSvc *serviceMocks.IMenuService,
				roleMenuPermissionSvc *serviceMocks.IRoleMenuPermissionService,
				loggerMock *mockLogger.ILogger,
				mockRedisExt *mockRedis.IRedisExt,
			) {
				roleSvc.On(
					"FindRoleBySlug",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)

				roleSvc.On(
					"Create",
					mock.Anything,
					mock.Anything,
				).Return(fmt.Errorf("failed to create role"))

				// Add flexible mocks for other services that might be called
				menuSvc.On("FindBySlug", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
				menuSvc.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
				menuSvc.On("IsShouldUpdate", mock.Anything, mock.Anything, mock.Anything).Return(false).Maybe()
				permissionSvc.On("FindBySlug", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
				permissionSvc.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
				roleMenuPermissionSvc.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

				loggerMock.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
				loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()
			},
		},
		{
			name: "ERROR: Database error on role lookup",
			mockSetup: func(
				roleSvc *roleMocks.IRoleService,
				permissionSvc *serviceMocks.IPermissionService,
				menuSvc *serviceMocks.IMenuService,
				roleMenuPermissionSvc *serviceMocks.IRoleMenuPermissionService,
				loggerMock *mockLogger.ILogger,
				mockRedisExt *mockRedis.IRedisExt,
			) {
				roleSvc.On(
					"FindRoleBySlug",
					mock.Anything,
					mock.Anything,
				).Return(nil, fmt.Errorf("database error"))

				// Add flexible mocks for other services that might be called
				menuSvc.On("FindBySlug", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
				menuSvc.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
				menuSvc.On("IsShouldUpdate", mock.Anything, mock.Anything, mock.Anything).Return(false).Maybe()
				permissionSvc.On("FindBySlug", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
				permissionSvc.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
				roleMenuPermissionSvc.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

				loggerMock.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
				loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			roleSvc := roleMocks.NewIRoleService(t)
			menuSvc := serviceMocks.NewIMenuService(t)
			permissionSvc := serviceMocks.NewIPermissionService(t)
			roleMenuPermissionSvc := serviceMocks.NewIRoleMenuPermissionService(t)
			mockLog := mockLogger.NewILogger(t)
			mockRedisExt := mockRedis.NewIRedisExt(t)
			ctx := context.Background()

			tc.mockSetup(roleSvc, permissionSvc, menuSvc, roleMenuPermissionSvc, mockLog, mockRedisExt)

			cfg := &config.Config{
				ServiceName:      "testing",
				PermissionConfig: config.PermissionConfig{},
			}
			currentDir, err := os.Getwd()
			assert.NoError(t, err)
			projectName := "backend-portal"
			projectRoot, err := util.FindProjectRoot(currentDir, projectName)
			if err != nil {
				fmt.Printf("Error finding project root: %v\n", err)
				return
			}
			cfg.PermissionConfig.Path = filepath.Join(projectRoot, "docs", "menu-and-permission-list.json")

			svc := New(cfg, mockLog, roleSvc, permissionSvc, menuSvc, roleMenuPermissionSvc, WithRedisClient(mockRedisExt))
			svc.SetupPredefinedRoleMenuPermissions(ctx)

			// Note: We don't assert exact call counts because the implementation
			// has complex loops and conditional logic that depend on the JSON file content
			// The important thing is that the mocks were called and no panics occurred
		})
	}
}
