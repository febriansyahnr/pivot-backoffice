package cancelMoneyFlowService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledgerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *CancelMoneyFlowService) CreateTransactions(ctx context.Context, request *ledgerModel.CreateNewLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/cancel/CreateTransactions")
	defer segment.End()

	var (
		ledgerTrx       []*orchestrator_model.AccountTransaction
		totalVoidAmount float64
	)

	trxList, err := s.repo.GetLedgerDetail(ctx, request.ReferenceID)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrGetLedgerDetail)
	}
	if len(trxList) == 0 {
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrLedgerDetailNotFound)
	}

	for _, trx := range trxList {
		if trx.SettlementStatus.String == constant.SettlementStatusPending && trx.AccountID == request.SenderAccountID {
			ledgerTrx = append(ledgerTrx, &trx)
		}

		if trx.AccountID == request.RecipientAccountID {
			totalVoidAmount += trx.Debit
		}
	}
	if len(ledgerTrx) == 0 {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrNotAllowedCancelTransaction)
	}
	request.Amount = totalVoidAmount

	ledgerTrx = ledgerModel.CancelTransactions(ledgerTrx)
	err = s.repo.BulkUpdateTransactions(ctx, ledgerTrx)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrUpdateLedgerEntry)
	}

	trx := ledgerModel.CreateNewCancelTransaction(request)
	err = s.repo.Create(ctx, trx)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrCreateLedgerEntry)
	}

	return nil
}
