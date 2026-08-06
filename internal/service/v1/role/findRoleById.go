package role

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *RoleService) FindRoleById(ctx context.Context, id string) (*role.Role, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/role/FindRoleById")
	defer segment.End()

	res, err := s.repo.FindRoleByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "error when finding role by id", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	if res == nil {
		s.logger.Error(ctx, fmt.Sprintf("role with id %s not found", id))
		return nil, errors.New(response.HttpErrNotFound, constant.ErrRoleNotFound)
	}

	return res, nil
}
