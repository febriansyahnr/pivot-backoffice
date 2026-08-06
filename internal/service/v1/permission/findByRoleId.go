package permissionService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PermissionService) FindByRoleId(ctx context.Context, roleId string) ([]*permissionModel.Permission, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/permission/FindByRoleId")
	defer segment.End()

	res, err := s.repo.FindByRoleId(ctx, roleId)
	if err != nil {
		return nil, pkgErrors.New(responseHttp.HttpErrDatabase, err)
	}

	// check if res is nil
	if res == nil {
		return nil, pkgErrors.New(responseHttp.HttpErrNotFound, constant.ErrPermissionNotFound)
	}

	return res, nil
}
