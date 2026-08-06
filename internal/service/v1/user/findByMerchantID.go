package user

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) ListUsersByMerchantID(
	ctx context.Context,
	filter *userModel.ListUsersByMerchantIDRequest,
	page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ListUsersByMerchantID")
	defer segment.End()

	err := filter.Validate()
	if err != nil {
		return nil, errors.New(response.HttpErrRequest, err)
	}

	users, err := s.userRepo.ListUsersByMerchantID(ctx, filter, page, perPage)
	if err != nil {
		s.logger.Error(ctx, "Failed to get list user", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	return users, nil
}
