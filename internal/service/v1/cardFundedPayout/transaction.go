package cardFundedPayoutService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) GetPayoutTransactionList(ctx context.Context, request model.GetPayoutTransactionListRequest) ([]model.GetPayoutTransactionListResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/GetPayoutTransactionList")
	defer span.End()

	transactions, err := s.disbursementRepo.GetCardFundedPayoutTransactionList(ctx, request)
	if err != nil {
		s.logger.Error(ctx, "Failed when get card funded payout transaction list", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	return transactions, nil
}
