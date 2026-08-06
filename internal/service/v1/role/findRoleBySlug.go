package role

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *RoleService) FindRoleBySlug(ctx context.Context, slug string) (*role.Role, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/role/FindRoleBySlug")
	defer segment.End()

	res, err := s.repo.FindRoleBySlug(ctx, slug)
	if err != nil {
		s.logger.Error(ctx, "error when finding role by slug", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	if res == nil {
		s.logger.Error(ctx, "role not found")
		return nil, errors.New(response.HttpErrNotFound, fmt.Errorf("role with slug %s not found", slug))
	}

	return res, nil
}
