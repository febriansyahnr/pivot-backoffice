package user

import (
	"context"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) FindUserByEmail(ctx context.Context, email string) (*userModel.User, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/FindUserByEmail")
	defer segment.End()

	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error(ctx, "error when finding user by email", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	return user, nil
}
