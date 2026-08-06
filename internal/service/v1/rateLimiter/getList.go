package ratelimiter

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *rateLimiterService) List(ctx context.Context, req *ratelimiter.MerchantRateLimitRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/rateLimiter/List")
	defer segment.End()

	list, total, err := s.rateLimiterRepo.List(ctx, req)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetRateLimiterConfigurationList)
	}

	return &commonModel.PaginationResponse{
		Data: list,
		Meta: *commonModel.NewMeta(req.Page, req.PageSize, total),
	}, nil
}
