package user

import (
	"context"
	"fmt"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) FindUserByID(ctx context.Context, id string) (*userModel.User, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/FindUserByID")
	defer segment.End()

	user, err := s.userRepo.FindUserByID(ctx, id)
	if err != nil {
		err = errors.New(response.HttpErrDatabase, err)
		s.logger.Error(ctx, "error when finding user by id", logger.Error(err))
		return nil, err
	}

	if user == nil {
		s.logger.Error(ctx, "user not found")
		return nil, errors.New(response.HttpErrNotFound, fmt.Errorf("user with id %s not found", id))
	}

	return user, nil
}
