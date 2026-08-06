package xbPayoutService

import (
	"context"

	payoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
)

func (s *xbPayoutService) GetXbPayoutDashboardInsights(ctx context.Context, request payoutModel.GetXbPayoutDashboardInsightRequest) (*payoutModel.XbPayoutDashboardInsights, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/GetXbPayoutDashboardInsights")
	defer segment.End()

	return s.disbursementRepo.GetXbPayoutDashboardInsights(ctx, request)
}
