package roleAndPermissionCronHandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"go.uber.org/zap"
)

func (h *RoleAndPermission) SetupPredefinedRoleMenuPermissions(ctx context.Context) {
	ctx, segment := otelTracer.Start(ctx, "cron/handler/roleAndPermission/SetupPredefinedRoleMenuPermissions")
	defer segment.End()

	// prepare predefined roles
	h.preparePredefinedRole(ctx)

	// read menu-and-permission-list from json file
	menuPermissionList, err := h.readMenuPermissionList(ctx)
	if err != nil {
		return
	}

	parentOrder := 0
	for _, menu := range menuPermissionList {
		parentMenu := h.saveMenuAndPermission(ctx, menu, nil, parentOrder)
		if parentMenu == nil {
			continue
		}
		parentOrder = parentMenu.Order + 1

		if len(menu.Children) > 0 {
			childOrder := 0
			for _, childMenu := range menu.Children {
				insertedChildMenu := h.saveMenuAndPermission(ctx, childMenu, parentMenu, childOrder)
				if insertedChildMenu != nil {
					childOrder = insertedChildMenu.Order + 1
				}
			}
		}
	}

}

func (h *RoleAndPermission) preparePredefinedRole(ctx context.Context) {
	// find or create predefined roles
	for _, predefinedRole := range constant.PredefinedRoles {
		// Find role by slug, skip on err
		existedRole, err := h.roleSvc.FindRoleBySlug(ctx, predefinedRole)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			h.logger.Info(ctx, err.Error(), zap.String("roleSlug", predefinedRole))
			continue
		}

		if existedRole != nil {
			// reset role permission cache first
			cacheKey := fmt.Sprintf(constant.PermissionByRoleKeyPattern, existedRole.UUID)
			if err = h.redisExt.Del(ctx, cacheKey).Err(); err != nil {
				continue
			}

		} else {
			// If existed role is not exist, then create one.
			existedRole = &role.Role{
				UUID:       uuid.NewString(),
				MerchantID: sql.NullString{String: "", Valid: false},
				Name:       util.ToTitle(predefinedRole),
				Slug:       predefinedRole,
				Type:       constant.RoleTypeDefault,
				CreatedAt:  time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
			}
			err = h.roleSvc.Create(ctx, existedRole)
			if err != nil {
				h.logger.Info(ctx, err.Error(), zap.Any("createRoleData", existedRole))
				continue
			}
		}
	}
}

func (h *RoleAndPermission) readMenuPermissionList(ctx context.Context) ([]roleMenuPermissionModel.MenuPermissionFromFileRequest, error) {
	// Open the JSON file
	file, err := os.Open(h.config.PermissionConfig.Path)
	if err != nil {
		h.logger.Error(ctx, "Error opening file", zap.Error(err))
		return nil, err
	}
	defer file.Close()

	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error(ctx, "Error read file", zap.Error(err))
		return nil, err
	}

	// Create a variable of the struct type to store the decoded data
	var data []roleMenuPermissionModel.MenuPermissionFromFileRequest

	// Unmarshal the JSON data into the struct
	err = json.Unmarshal(content, &data)
	if err != nil {
		h.logger.Error(ctx, "Error unmarshalling JSON", zap.Error(err))
		return nil, err
	}

	return data, nil
}

func (h *RoleAndPermission) saveMenuAndPermission(
	ctx context.Context,
	menu roleMenuPermissionModel.MenuPermissionFromFileRequest,
	parentMenu *menuModel.Menu,
	order int,
) *menuModel.Menu {
	var (
		level = 0
	)
	if parentMenu != nil {
		level = parentMenu.Level + 1
	}

	// Get menu by slug
	existedMenu, err := h.menuSvc.FindBySlug(ctx, menu.Slug)
	if err != nil && !errors.Is(err, constant.ErrMenuNotFound) {
		h.logger.Info(ctx, err.Error(), zap.String("menuSlug", menu.Slug))
		return nil
	}

	if existedMenu == nil {
		// Build menu
		existedMenu = &menuModel.Menu{
			UUID:      uuid.NewString(),
			Slug:      menu.Slug,
			Name:      menu.Name,
			Type:      menu.Type,
			Icon:      menu.Icon,
			Path:      menu.Path,
			Level:     level,
			Order:     order,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if parentMenu != nil {
			existedMenu.ParentID = &parentMenu.UUID
		}

		// Create menu
		if err = h.menuSvc.Create(ctx, existedMenu); err != nil {
			h.logger.Info(ctx, err.Error(), zap.Any("menuData", existedMenu))
			return nil
		}
	}

	// Update the menu when changeable fields are different
	if h.menuSvc.IsShouldUpdate(ctx, existedMenu, menu) {
		existedMenu.Name = menu.Name
		existedMenu.Icon = menu.Icon

		if err = h.menuSvc.Update(ctx, existedMenu); err != nil {
			h.logger.Info(ctx, "err-update-menu : "+err.Error(), zap.Any("menuData", existedMenu))
			return nil
		}
	}

	// Loop permissions
	for _, permission := range menu.Permissions {
		// find or create permission
		existedPermission, err := h.permissionSvc.FindBySlug(ctx, permission.Slug)
		if err != nil && !errors.Is(err, constant.ErrPermissionNotFound) {
			h.logger.Info(ctx, err.Error(), zap.String("permissionSlug", permission.Slug))
			continue
		}

		if existedPermission == nil {
			existedPermission = &permissionModel.Permission{
				UUID:        uuid.NewString(),
				Slug:        permission.Slug,
				Name:        permission.Name,
				Description: permission.Description,
				Group:       permission.Group,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			}

			err = h.permissionSvc.Create(ctx, existedPermission)
			if err != nil {
				h.logger.Info(ctx, err.Error(), zap.Any("createPermissionData", existedPermission))
				continue
			}
		} else {
			existedPermission.Slug = permission.Slug
			existedPermission.Name = permission.Name
			existedPermission.Description = permission.Description
			existedPermission.Group = permission.Group
			existedPermission.UpdatedAt = time.Now().UTC()

			if err = h.permissionSvc.Update(ctx, existedPermission); err != nil {
				h.logger.Info(ctx, "err-update-permission : "+err.Error(), zap.Any("permissionData", existedPermission))
				continue
			}
		}

		for _, roleSlug := range permission.Roles {
			// get role by slug
			existedRole, err := h.roleSvc.FindRoleBySlug(ctx, roleSlug)
			if err != nil && strings.Contains(err.Error(), "not found") {
				h.logger.Info(ctx, err.Error(), zap.String("roleSlug", roleSlug))
				continue
			}

			if existedRole == nil {
				h.logger.Info(ctx, "Role is not in predefined role list", zap.String("roleSlug", roleSlug))
				continue
			}

			// save to role menu permission pivot table
			if err = h.roleMenuPermissionSvc.Create(ctx, &roleMenuPermissionModel.RoleMenuPermission{
				RoleID:       existedRole.UUID,
				MenuID:       existedMenu.UUID,
				PermissionID: existedPermission.UUID,
			}); err != nil {
				h.logger.Error(ctx, err.Error(), zap.Error(err))
				continue
			}
		}
	}

	return existedMenu
}
