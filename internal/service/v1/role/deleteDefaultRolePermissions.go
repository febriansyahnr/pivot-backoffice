package role

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/exp/maps"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *RoleService) DeleteDefaultRolePermissions(ctx context.Context, payload *role.CRMUpdateDefaultRolePermissionsRequest) (resp *role.RoleMenuResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/role/DeleteDefaultRolePermissions")
	defer segment.End()

	// Find the default role by slug
	existedRole, err := s.FindRoleBySlug(ctx, payload.RoleSlug)
	if err != nil {
		return nil, err
	}

	// Validate that this is a default role
	if existedRole.Type != constant.RoleTypeDefault {
		return nil, pkgErrs.New(response.HttpErrRequest, errors.New("role is not a default role"))
	}

	// Validate that merchant_id is NULL for default roles
	if existedRole.MerchantID.Valid {
		return nil, pkgErrs.New(response.HttpErrRequest, errors.New("default role should not have merchant_id"))
	}

	// Clear cache for this role
	cacheKey := fmt.Sprintf(constant.PermissionByRoleKeyPattern, existedRole.UUID)
	_ = s.redis.Del(ctx, cacheKey)

	// Process request menus and permissions
	requestedMenuPermissions := map[string]map[string]bool{}
	for _, menu := range payload.Menus {
		if _, ok := requestedMenuPermissions[menu.Slug]; !ok {
			requestedMenuPermissions[menu.Slug] = map[string]bool{}
		}
		for _, perm := range menu.Permissions {
			requestedMenuPermissions[menu.Slug][perm] = true
		}
	}

	// Validate all requested menus and permissions exist
	menuPermissionIDs := make(map[string]map[string]string) // menuSlug -> permissionSlug -> permissionID
	menuIDs := make(map[string]string)                      // menuSlug -> menuID

	for menuSlug, permissions := range requestedMenuPermissions {
		result, err := s.menuRepo.GetMenuAndPermissionIDs(ctx, menuSlug, maps.Keys(permissions)...)
		if err != nil {
			s.logger.Error(ctx, "failed to get menu and permission IDs",
				logger.Error(err),
				logger.String("menuSlug", menuSlug),
				logger.String("roleSlug", payload.RoleSlug))
			return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrGetMenuPermission)
		}
		if result == nil || len(result.Permissions) != len(permissions) {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMenuOrPermissionNotRegistered)
		}

		menuIDs[menuSlug] = result.MenuID
		menuPermissionIDs[menuSlug] = make(map[string]string)

		// Build a map from the requested permission slugs to their IDs
		permSlugList := maps.Keys(permissions)
		for i, perm := range result.Permissions {
			if i < len(permSlugList) {
				menuPermissionIDs[menuSlug][permSlugList[i]] = perm.ID
			}
		}
	}

	// Delete specified permissions
	for menuSlug, permissions := range requestedMenuPermissions {
		menuID := menuIDs[menuSlug]
		permIDs := []string{}
		for permSlug := range permissions {
			permIDs = append(permIDs, menuPermissionIDs[menuSlug][permSlug])
		}

		if err = s.roleMenuPermRepo.DeleteByMenuAndPermissions(ctx, existedRole.UUID, menuID, permIDs); err != nil {
			s.logger.Error(ctx, "failed to delete role menu permissions",
				logger.Error(err),
				logger.String("roleSlug", payload.RoleSlug),
				logger.String("roleId", existedRole.UUID),
				logger.String("menuSlug", menuSlug),
				logger.String("menuId", menuID))
			return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrDeleteRoleMenuPermissions)
		}
	}

	// Update role timestamp
	roleData := &role.Role{
		UUID:       existedRole.UUID,
		MerchantID: existedRole.MerchantID,
		Name:       existedRole.Name,
		Slug:       existedRole.Slug,
		Type:       existedRole.Type,
		CreatedAt:  existedRole.CreatedAt,
		UpdatedAt:  time.Now().UTC(),
		DeletedAt:  existedRole.DeletedAt,
	}

	if err = s.Update(ctx, roleData); err != nil {
		s.logger.Error(ctx, "failed to update role timestamp",
			logger.Error(err),
			logger.String("roleSlug", payload.RoleSlug),
			logger.String("roleId", existedRole.UUID))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrUpdateRoleData)
	}

	// Get the current state after operation to return
	currentMenus, err := s.menuRepo.GetAll(ctx, &menuModel.GetAllFilterRequest{RoleID: existedRole.UUID})
	if err != nil {
		s.logger.Warn(ctx, "Failed to get current menus after operation", logger.Error(err))
		currentMenus = []*menuModel.MenuResponse{}
	}

	// Build response
	resp = &role.RoleMenuResponse{
		ID:    roleData.UUID,
		Name:  existedRole.Name,
		Menus: []role.RoleMenuPermissionResponse{},
	}

	for _, menu := range currentMenus {
		permSlugList := make([]string, len(menu.Permissions))
		for i, perm := range menu.Permissions {
			permSlugList[i] = perm.Slug
		}
		resp.Menus = append(resp.Menus, role.RoleMenuPermissionResponse{
			Name:        menu.Name,
			Permissions: permSlugList,
		})
	}

	s.logger.Info(ctx, "Default role permissions deleted successfully",
		logger.String("roleSlug", payload.RoleSlug),
		logger.String("roleId", existedRole.UUID))

	return
}
