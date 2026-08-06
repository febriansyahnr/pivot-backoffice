package role

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *RoleService) GetList(
	ctx context.Context,
	filter *role.GetRoleFilterRequest,
	page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/role/GetList")
	defer segment.End()

	roles, err := s.repo.GetList(ctx, filter, page, perPage)
	if err != nil {
		s.logger.Error(ctx, "Failed to get list role", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	return roles, nil
}
