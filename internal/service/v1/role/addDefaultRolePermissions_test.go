package role

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mockMenu "github.com/paper-indonesia/pivot-backoffice/mocks/repository/menu"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/repository/role"
	mockRoleMenuPerm "github.com/paper-indonesia/pivot-backoffice/mocks/repository/roleMenuPermission"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddDefaultRolePermissions(t *testing.T) {
	now := time.Now()

	defaultRole := &role.Role{
		UUID:       "role-uuid-123",
		MerchantID: sql.NullString{Valid: false},
		Name:       "Admin",
		Slug:       "ADMIN",
		Type:       constant.RoleTypeDefault,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	customRole := &role.Role{
		UUID:       "role-uuid-456",
		MerchantID: sql.NullString{String: "merchant-123", Valid: true},
		Name:       "Custom Role",
		Slug:       "CUSTOM",
		Type:       constant.RoleTypeCustom,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	testCases := []struct {
		Name          string
		IsSuccess     bool
		Payload       *role.CRMUpdateDefaultRolePermissionsRequest
		ExpectedError string
		MockSetup     func(
			roleRepo *mockRole.IRoleRepository,
			menuRepo *mockMenu.IMenuRepository,
			roleMenuPermRepo *mockRoleMenuPerm.IRoleMenuPermissionRepository,
			redisClient *mockRedis.IRedisExt,
		)
	}{
		{
			Name:      "SUCCESS: add permissions to default role",
			IsSuccess: true,
			Payload: &role.CRMUpdateDefaultRolePermissionsRequest{
				RoleSlug: "ADMIN",
				Menus: []role.RoleMenuRequest{
					{
						Slug:        "payment",
						Permissions: []string{"payment.link.create", "payment.link.read"},
					},
				},
			},
			MockSetup: func(roleRepo *mockRole.IRoleRepository, menuRepo *mockMenu.IMenuRepository, roleMenuPermRepo *mockRoleMenuPerm.IRoleMenuPermissionRepository, redisClient *mockRedis.IRedisExt) {
				roleRepo.On("FindRoleBySlug", mock.Anything, "ADMIN").Return(defaultRole, nil)
				redisClient.On("Del", mock.Anything, mock.AnythingOfType("string")).Return(nil)
				menuRepo.On("GetMenuAndPermissionIDs", mock.Anything, "payment", mock.Anything, mock.Anything).
					Return(&menuModel.MenuAndPermissionIDs{
						MenuID:   "menu-uuid-1",
						MenuName: "Accept Payments",
						Permissions: []menuModel.ObjPermission{
							{ID: "perm-uuid-1", Name: "Create Payment Link"},
							{ID: "perm-uuid-2", Name: "Read Payment Link"},
						},
					}, nil)
				roleMenuPermRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Times(2)
				roleRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
				menuRepo.On("GetAll", mock.Anything, &menuModel.GetAllFilterRequest{RoleID: defaultRole.UUID}).
					Return([]*menuModel.MenuResponse{
						{
							UUID: "menu-uuid-1",
							Name: "Accept Payments",
							Slug: "payment",
							Permissions: []menuModel.MenuPermission{
								{Slug: "payment.link.create"},
								{Slug: "payment.link.read"},
							},
						},
					}, nil)
			},
		},
		{
			Name:      "ERROR: role not found",
			IsSuccess: false,
			Payload: &role.CRMUpdateDefaultRolePermissionsRequest{
				RoleSlug: "NONEXISTENT",
				Menus: []role.RoleMenuRequest{
					{
						Slug:        "payment",
						Permissions: []string{"payment.link.create"},
					},
				},
			},
			ExpectedError: "not found",
			MockSetup: func(roleRepo *mockRole.IRoleRepository, menuRepo *mockMenu.IMenuRepository, roleMenuPermRepo *mockRoleMenuPerm.IRoleMenuPermissionRepository, redisClient *mockRedis.IRedisExt) {
				roleRepo.On("FindRoleBySlug", mock.Anything, "NONEXISTENT").Return(nil, errors.New("role not found"))
			},
		},
		{
			Name:      "ERROR: role is not a default role",
			IsSuccess: false,
			Payload: &role.CRMUpdateDefaultRolePermissionsRequest{
				RoleSlug: "CUSTOM",
				Menus: []role.RoleMenuRequest{
					{
						Slug:        "payment",
						Permissions: []string{"payment.link.create"},
					},
				},
			},
			ExpectedError: "role is not a default role",
			MockSetup: func(roleRepo *mockRole.IRoleRepository, menuRepo *mockMenu.IMenuRepository, roleMenuPermRepo *mockRoleMenuPerm.IRoleMenuPermissionRepository, redisClient *mockRedis.IRedisExt) {
				roleRepo.On("FindRoleBySlug", mock.Anything, "CUSTOM").Return(customRole, nil)
			},
		},
		{
			Name:      "ERROR: default role should not have merchant_id",
			IsSuccess: false,
			Payload: &role.CRMUpdateDefaultRolePermissionsRequest{
				RoleSlug: "ADMIN",
				Menus: []role.RoleMenuRequest{
					{
						Slug:        "payment",
						Permissions: []string{"payment.link.create"},
					},
				},
			},
			ExpectedError: "default role should not have merchant_id",
			MockSetup: func(roleRepo *mockRole.IRoleRepository, menuRepo *mockMenu.IMenuRepository, roleMenuPermRepo *mockRoleMenuPerm.IRoleMenuPermissionRepository, redisClient *mockRedis.IRedisExt) {
				invalidRole := &role.Role{
					UUID:       "role-uuid-789",
					MerchantID: sql.NullString{String: "merchant-123", Valid: true},
					Name:       "Admin",
					Slug:       "ADMIN",
					Type:       constant.RoleTypeDefault,
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				roleRepo.On("FindRoleBySlug", mock.Anything, "ADMIN").Return(invalidRole, nil)
			},
		},
		{
			Name:      "ERROR: menu or permission not registered",
			IsSuccess: false,
			Payload: &role.CRMUpdateDefaultRolePermissionsRequest{
				RoleSlug: "ADMIN",
				Menus: []role.RoleMenuRequest{
					{
						Slug:        "payment",
						Permissions: []string{"payment.invalid.permission"},
					},
				},
			},
			ExpectedError: "menu or permission not registered",
			MockSetup: func(roleRepo *mockRole.IRoleRepository, menuRepo *mockMenu.IMenuRepository, roleMenuPermRepo *mockRoleMenuPerm.IRoleMenuPermissionRepository, redisClient *mockRedis.IRedisExt) {
				roleRepo.On("FindRoleBySlug", mock.Anything, "ADMIN").Return(defaultRole, nil)
				redisClient.On("Del", mock.Anything, mock.AnythingOfType("string")).Return(nil)
				menuRepo.On("GetMenuAndPermissionIDs", mock.Anything, "payment", "payment.invalid.permission").
					Return(nil, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			roleRepo := mockRole.NewIRoleRepository(t)
			menuRepo := mockMenu.NewIMenuRepository(t)
			roleMenuPermRepo := mockRoleMenuPerm.NewIRoleMenuPermissionRepository(t)
			redisClient := mockRedis.NewIRedisExt(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.MockSetup(roleRepo, menuRepo, roleMenuPermRepo, redisClient)

			svc := New(
				roleRepo,
				loggerMock,
				WithMenuRepository(menuRepo),
				WithRoleMenuPermissionRepository(roleMenuPermRepo),
				WithRedisClient(redisClient),
			)

			response, err := svc.AddDefaultRolePermissions(context.Background(), tc.Payload)

			if tc.IsSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, defaultRole.UUID, response.ID)
				assert.Equal(t, defaultRole.Name, response.Name)
				assert.NotEmpty(t, response.Menus)
			} else {
				assert.Error(t, err)
				assert.Nil(t, response)
				assert.Contains(t, err.Error(), tc.ExpectedError)
			}

			roleRepo.AssertExpectations(t)
			menuRepo.AssertExpectations(t)
			roleMenuPermRepo.AssertExpectations(t)
			redisClient.AssertExpectations(t)
		})
	}
}
