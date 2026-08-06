package paymentService

import (
	"context"
	"time"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"

	"golang.org/x/sync/errgroup"
)

func (s *PaymentService) GetPaymentDashboardInsights(ctx context.Context, request model.GetPaymentDashboardInsightRequest) (result *model.PaymentDashboardInsights, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetPaymentDashboardInsights")
	defer segment.End()

	var (
		currentStartDate = request.StartDate.In(loc)
		currentEndDate   = request.EndDate.In(loc)
		prevStartDate    = currentStartDate.AddDate(0, 0, -int((currentEndDate.Sub(currentStartDate).Hours()/24)+1))
		prevEndDate      = currentStartDate.Add(-time.Second)
		successRate      = &model.PaymentSuccessRateComparison{}
	)

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() (er error) {
		result, er = s.paymentRepo.GetPaymentDashboardInsights(ctx, request)

		return er
	})

	errGroup.Go(func() (er error) {
		successRate, er = s.paymentMetricsRepo.QueryPaymentSuccessRateComparison(ctx, model.QueryPaymentSuccessRateComparisonRequest{
			MerchantId: request.MerchantId,
			PrevRange: model.DateRange{
				StartDate: prevStartDate.Format(time.DateOnly),
				EndDate:   prevEndDate.Format(time.DateOnly),
			},
			CurrentRange: model.DateRange{
				StartDate: currentStartDate.Format(time.DateOnly),
				EndDate:   currentEndDate.Format(time.DateOnly),
			},
		})
		return er
	})

	if err = errGroup.Wait(); err != nil {
		return nil, err
	}

	result.SuccessRate = successRate

	return
}
