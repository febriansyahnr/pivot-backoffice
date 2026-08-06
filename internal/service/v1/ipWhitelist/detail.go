package ipwhitelistService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *IPWhitelistService) Detail(ctx context.Context, merchantId, uuid string) (*ipwhitelistModel.IPWhitelistConfiguration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ipWhitelist/Detail")
	defer segment.End()

	config, err := s.whitelistRepo.Detail(ctx, uuid)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetIPWhitelistConfigurationDetail)
	}
	if config == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrIPWhitelistConfigurationNotFound)
	}
	if config.MerchantID != merchantId {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidIPConfigurationID)
	}

	return config, nil
}
