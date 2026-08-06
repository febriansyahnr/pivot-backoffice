package ipwhitelistService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *IPWhitelistService) Update(ctx context.Context, req *ipwhitelistModel.UpdateIPWhitelistConfiguration) (*ipwhitelistModel.IPWhitelistConfiguration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ipWhitelist/Update")
	defer segment.End()

	// Pending validation of ip & subnet
	config, err := s.whitelistRepo.Detail(ctx, req.ID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetIPWhitelistConfigurationDetail)
	}
	if config == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrIPWhitelistConfigurationNotFound)
	}

	err = config.Update(req)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}
	err = s.whitelistRepo.Update(ctx, config)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrUpdateIPWhitelistConfiguration)
	}

	err = s.updateCache(ctx, req.MerchantID)
	if err != nil {
		return nil, err
	}

	return config, nil
}
