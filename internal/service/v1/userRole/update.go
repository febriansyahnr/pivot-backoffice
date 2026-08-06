package userRole

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserRoleService) UpdateByUserID(ctx context.Context, userRole *userRole.UserRole) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/userRole/UpdateByUserID")
	defer segment.End()

	if err := s.repo.UpdateByUserID(ctx, userRole); err != nil {
		s.logger.Error(ctx, "Failed to update user_role", logger.Error(err))
		return errors.New(response.HttpErrDatabase, err)
	}

	return nil
}
