package role

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/exp/maps"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *RoleService) Update(ctx context.Context, role *role.Role) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/role/Update")
	defer segment.End()

	if err := s.repo.Update(ctx, role); err != nil {
		s.logger.Error(ctx, "Failed to update role", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	return nil
}

func (s *RoleService) UpdateRoleAndPermissions(ctx context.Context, payload *role.UpdateRoleRequest) (resp *role.RoleMenuResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/role/CreateRoleAndPermissions")
	defer segment.End()

	// check if role exists
	existedRole, err := s.FindRoleById(ctx, payload.RoleID)
	if err != nil {
		return nil, err
	}

	if existedRole.Type == constant.RoleTypeDefault {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrCannotModifyDefaultRole)
	}

	cacheKey := fmt.Sprintf(constant.PermissionByRoleKeyPattern, payload.RoleID)
	_ = s.redis.Del(ctx, cacheKey)

	// Eliminate data duplication
	menuPermissions := map[string]map[string]bool{}
	for _, menu := range payload.Menus {
		if _, ok := menuPermissions[menu.Slug]; !ok {
			menuPermissions[menu.Slug] = map[string]bool{}
		}
		for _, perm := range menu.Permissions {
			menuPermissions[menu.Slug][perm] = true
		}
	}
	for _, combinations := range combinationsAreNotAllowed {
		exists := 0
		for _, menuName := range combinations {
			if _, ok := menuPermissions[menuName]; ok {
				exists++
			}
		}
		if len(combinations) == exists {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("combination of menu access is not allowed"))
		}
	}

	roleData := &role.Role{
		UUID:       existedRole.UUID,
		MerchantID: existedRole.MerchantID,
		Name:       payload.Name,
		Slug:       existedRole.Slug,
		Type:       existedRole.Type,
		CreatedAt:  existedRole.CreatedAt,
		UpdatedAt:  time.Now().UTC(),
		DeletedAt:  existedRole.DeletedAt,
	}

	// Response function
	resp = &role.RoleMenuResponse{
		ID:    roleData.UUID,
		Name:  payload.Name,
		Menus: []role.RoleMenuPermissionResponse{},
	}
	roleMenuPermissions := []*roleMenuPermissionModel.RoleMenuPermission{}

	for menu, permissions := range menuPermissions {
		result, err := s.menuRepo.GetMenuAndPermissionIDs(ctx, menu, maps.Keys(permissions)...)
		if err != nil {
			return nil, err

		} else if result == nil || len(result.Permissions) != len(permissions) {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("menu or permission not registered"))
		}

		resp.Menus = append(resp.Menus, role.RoleMenuPermissionResponse{
			Name:        result.MenuName,
			Permissions: make([]string, len(result.Permissions)),
		})
		for i, perm := range result.Permissions {
			roleMenuPermissions = append(roleMenuPermissions, &roleMenuPermissionModel.RoleMenuPermission{
				RoleID:       roleData.UUID,
				MenuID:       result.MenuID,
				PermissionID: perm.ID,
			})
			resp.Menus[len(resp.Menus)-1].Permissions[i] = perm.Name
		}
	}

	if ctx, err = s.repo.BeginTransaction(ctx); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.repo.RollbackTransaction(ctx); e != nil {
				resp, err = nil, pkgErrs.New(response.HttpErrDatabase, e)
			}
		}
	}()

	// fill role data
	if err = s.Update(ctx, roleData); err != nil {
		return nil, err
	}

	// delete menu permission by role
	if errDelete := s.roleMenuPermRepo.Delete(ctx, roleData.UUID); errDelete != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	// re create role menu permission
	for _, roleMenu := range roleMenuPermissions {
		if err = s.roleMenuPermRepo.Create(ctx, roleMenu); err != nil {
			return nil, pkgErrs.New(response.HttpErrDatabase, err)
		}
	}

	if err = s.repo.CommitTransaction(ctx); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	isCompleted = true

	return
}
