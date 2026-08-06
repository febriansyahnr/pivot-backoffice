package paymentService

import (
	"context"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
)

func (s *PaymentService) GetInvestigationSummary(ctx context.Context, opt paymentModel.GetInvestigationSummaryOption) (*paymentModel.InvestigationSummaryResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetInvestigationSummary")
	defer segment.End()

	return s.paymentRepo.GetInvestigationSummary(ctx, opt)
}
