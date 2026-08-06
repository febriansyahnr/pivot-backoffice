package roleMenuPermissionService

import (
	"context"

	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
)

func (s *RoleMenuPermissionService) Create(ctx context.Context, pivot *roleMenuPermissionModel.RoleMenuPermission) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/roleMenuPermission/Create")
	defer segment.End()

	return s.repo.Create(ctx, pivot)
}
