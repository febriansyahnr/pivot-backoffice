package ipwhitelistService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *IPWhitelistService) List(ctx context.Context, req *ipwhitelistModel.GetIPWhitelistConfiguration) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ipWhitelist/List")
	defer segment.End()

	list, total, err := s.whitelistRepo.List(ctx, req)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetIPWhitelistConfigurationList)
	}

	return &commonModel.PaginationResponse{
		Data: list,
		Meta: *commonModel.NewMeta(req.Page, req.PageSize, total),
	}, nil
}
