package xbPayoutService

import (
	"context"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
)

func (s *xbPayoutService) GetFeeConfig(ctx context.Context, merchantID string) (*merchantModel.XbFeeConfigResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetFeeConfig")
	defer segment.End()

	return s.feeSvc.GetXbFeeConfigs(ctx, merchantID)
}
