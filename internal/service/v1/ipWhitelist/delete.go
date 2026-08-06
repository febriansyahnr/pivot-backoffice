package ipwhitelistService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *IPWhitelistService) Delete(ctx context.Context, merchantId, uuid string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ipWhitelist/Delete")
	defer segment.End()

	config, err := s.whitelistRepo.Detail(ctx, uuid)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrGetIPWhitelistConfigurationDetail)
	}
	if config == nil {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrIPWhitelistConfigurationNotFound)
	}
	if config.MerchantID != merchantId {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidIPConfigurationID)
	}

	err = s.whitelistRepo.Delete(ctx, uuid)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrDeleteIPWhitelistConfiguration)
	}

	err = s.updateCache(ctx, merchantId)
	if err != nil {
		return err
	}

	return nil
}
