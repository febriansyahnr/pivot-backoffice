package xbPayoutService

import (
	"context"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *xbPayoutService) GetFxRate(ctx context.Context, request *xbModel.GetFxRateRequest) (*xbModel.GetFxRateResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetFxRate")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    "",
		ReferenceId: request.MerchantId,
		From:        serviceName,
	})

	xbCoreResp, err := s.xbCoreProcessorRepo.GetFxRate(ctx, &xbCoreProcessorModel.GetFxRateRequest{
		MerchantId:          request.MerchantId,
		SourceCurrency:      request.SourceCurrency,
		DestinationCurrency: request.DestinationCurrency,
	})
	if err != nil {
		s.logger.Error(ctx, "GetFxRate - Failed to get FX rate from xb-core-processor", logger.Error(err))
		return nil, err
	}

	return &xbModel.GetFxRateResponse{
		FxRate:            xbCoreResp.MarkupFxRate,
		DestinationFxRate: xbCoreResp.DestinationFxRate,
		ExpiryAt:          xbCoreResp.ExpiryAt,
	}, nil
}
