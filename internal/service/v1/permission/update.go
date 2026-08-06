package permissionService

import (
	"context"

	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
)

func (s *PermissionService) Update(ctx context.Context, permission *permissionModel.Permission) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/permission/Update")
	defer segment.End()

	return s.repo.Update(ctx, permission)
}
