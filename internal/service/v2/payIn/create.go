package payInMoneyFlowService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PayInMoneyFlowService) CreateTransactions(ctx context.Context, request *ledger_model.CreateNewLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/payIn/CreateTransactions")
	defer segment.End()

	err := request.ValidatePayInRequest()
	if err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	trxList, err := ledger_model.CreatePayInTransactions(ctx, request)
	if err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	err = s.repo.BulkInsert(ctx, trxList)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrCreateLedgerEntry)
	}

	return nil
}
