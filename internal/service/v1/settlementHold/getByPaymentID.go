package settlementHoldService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *settlementHoldService) GetSettlementHoldByPaymentID(ctx context.Context, paymentID string) (*settlementHold.SettlementHold, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/settlementHold/GetSettlementHoldByPaymentID")
	defer segment.End()

	settlementHoldRecord, err := s.repo.GetByPaymentID(ctx, paymentID)
	if err != nil {
		s.logger.Error(ctx, "error when get settlement hold by payment id", logger.Error(err), logger.String("paymentId", paymentID))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetSettlementHold)
	}

	return settlementHoldRecord, nil
}
