package user

import (
	"context"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) Create(ctx context.Context, user *userModel.User) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/Create")
	defer segment.End()

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to create user", logger.Error(err))
		return errors.New(response.HttpErrDatabase, err)
	}

	return nil
}
