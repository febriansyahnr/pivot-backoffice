package refundMoneyFlowService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledgerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Reversal or Refund case
func (s *RefundMoneyFlowService) CreateTransactions(ctx context.Context, request *ledgerModel.CreateNewLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/refund/CreateTransactions")
	defer segment.End()

	trxList, err := ledgerModel.NewRefundTransactions(request)
	if err != nil {
		s.logger.Error(ctx, "error when create refund transactions", logger.Error(err), logger.Any("request", request))
		return pkgErrors.New(response.HttpErrInternal, constant.ErrCreateLedgerEntry)
	}
	err = s.repo.BulkInsert(ctx, trxList)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrStoreLedgerEntry)
	}

	return nil
}
