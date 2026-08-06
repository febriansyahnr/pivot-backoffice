package role

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/exp/maps"
)

func (s *RoleService) Create(ctx context.Context, role *role.Role) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/role/Create")
	defer segment.End()

	if err := s.repo.Create(ctx, role); err != nil {
		s.logger.Error(ctx, "Failed to create role", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, err)
	}

	return nil
}

func (s *RoleService) CreateRoleAndPermissions(ctx context.Context, payload *role.CreateRoleRequest) (resp *role.RoleMenuResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/role/CreateRoleAndPermissions")
	defer segment.End()

	if total, err := s.repo.TotalRoleByMerchantID(ctx, payload.MerchantID); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if total >= constant.MaxRolesPerMerchant {
		return nil, pkgErrs.New(response.HttpErrTooManyRequest, errors.New("role limit have been exceeded"))
	}

	if avail, err := s.repo.CheckAvailableRoleName(ctx, payload.MerchantID, payload.Name); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if !avail {
		return nil, pkgErrs.New(response.HttpErrDupCheck, errors.New("role name is already in use"))
	}

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
		UUID:      uuid.New().String(),
		Name:      payload.Name,
		Slug:      fmt.Sprintf("%s:%d", util.CreateSlug(payload.Name), time.Now().Unix()),
		Type:      constant.RoleTypeCustom,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		MerchantID: sql.NullString{
			Valid: true, String: payload.MerchantID,
		},
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

	if err = s.repo.Create(ctx, roleData); err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
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
