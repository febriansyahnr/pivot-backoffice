package cardFundedPayoutService

import (
	"context"

	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
)

func (s *service) GetPayoutDetail(
	ctx context.Context,
	request *cardFundedPayoutModel.GetPayoutDetailRequest,
) (*cardFundedPayoutModel.GetPayoutDetailResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/GetPayoutDetail")
	defer span.End()

	// Get payout detail from repository
	resp, err := s.disbursementRepo.GetCardFundedPayoutDetail(ctx, request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
