package user

import (
	"context"
	"fmt"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// ListUsers is a function to handle the get user list request
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*userModel.User, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ListUsers")
	defer segment.End()

	users, err := s.userRepo.ListUsers(ctx, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to get user list", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	if len(users) == 0 {
		s.logger.Error(ctx, "User not found")
		return nil, errors.New(response.HttpErrNotFound, fmt.Errorf("user not found"))
	}

	return users, nil
}
