package payoutMoneyFlowService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PayOutMoneyFlowService) CreateTransactions(ctx context.Context, request *ledger_model.CreateNewLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/payOut/CreateTransactions")
	defer segment.End()

	err := request.ValidatePayOutRequest()
	if err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	balance, err := s.ledgerSvc.GetLedgerBalance(ctx, request.SenderAccountID)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrFailedValidatePayOut)
	}
	if balance.Balance < request.Amount+request.Fee.Amount {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrInsufficientBalance)
	}

	trxList, err := ledger_model.CreatePayOutTransactions(ctx, request)
	if err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	err = s.repo.BulkInsert(ctx, trxList)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrCreateLedgerEntry)
	}

	return nil
}
