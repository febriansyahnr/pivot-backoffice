package cardFundedPayoutService

import (
	"context"

	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

func (s *service) GetPayoutList(
	ctx context.Context,
	filter *cardFundedPayoutModel.FilterGetPayoutList,
) (*commonModel.PaginationResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/GetPayoutList")
	defer span.End()

	resp, err := s.disbursementRepo.GetCardFundedPayoutList(ctx, filter)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
