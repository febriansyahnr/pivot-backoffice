package cardFundedPayoutService

import (
	"context"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

// GetPayoutInsights returns the total amount and transaction count
// for card-funded payouts with WAITING approval status.
func (s *service) GetPayoutInsights(
	ctx context.Context,
	filter *cardFundedPayoutModel.FilterGetPayoutInsights,
) (*cardFundedPayoutModel.GetPayoutInsightsResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/GetPayoutInsights")
	defer span.End()

	dto, err := s.disbursementRepo.GetCardFundedPayoutInsights(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &cardFundedPayoutModel.GetPayoutInsightsResponse{
		TotalTransaction: dto.Count,
		TotalAmount: commonModel.Amount{
			Currency: constant.CurrencyIDR,
			Value:    strconv.FormatFloat(dto.Sum, 'f', 2, 64),
		},
	}, nil
}
