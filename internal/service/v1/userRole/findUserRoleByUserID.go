package userRole

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserRoleService) FindUserRoleByUserID(ctx context.Context, userId string) (*userRole.UserRole, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/userRole/FindUserRoleByUserIDs")
	defer segment.End()

	res, err := s.repo.FindUserRoleByUserID(ctx, userId)
	if err != nil {
		s.logger.Error(ctx, "error when finding user_role by user_id", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	if res == nil {
		s.logger.Error(ctx, "user_role not found")
		return nil, errors.New(response.HttpErrNotFound, fmt.Errorf("user_role with user_id %s not found", userId))
	}

	return res, nil
}
