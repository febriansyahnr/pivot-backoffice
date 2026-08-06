package permissionService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PermissionService) FindBySlug(ctx context.Context, slug string) (*permissionModel.Permission, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/permission/FindBySlug")
	defer segment.End()

	res, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if res == nil {
		err = constant.ErrPermissionNotFound
		return nil, pkgErrors.New(response.HttpErrNotFound, err)
	}

	return res, nil
}
