package ledgerService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *LedgerService) RecordTransaction(ctx context.Context, merchantId string, request *ledger_model.CreateNewLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/ledger/RecordTransaction")
	defer segment.End()

	err := s.validatorSvc.ValidateTransaction(ctx, merchantId, request)
	if err != nil {
		return err
	}

	if s.moneyFlowSvcMap == nil {
		return pkgErr.New("service not initialized", fmt.Errorf("moneyFlowSvcMap is nil"))
	}
	moneyFlowSvc, exists := s.moneyFlowSvcMap[request.TransferType]
	if !exists {
		return pkgErr.New("invalid transfer type", fmt.Errorf("transfer type %s not supported", request.TransferType))
	}

	isTrxCompleted := false
	existingTrxCtx := ctx.Value(pdkConstant.CtxSqlTx)
	if existingTrxCtx == nil {
		ctx, err := s.repo.BeginTransaction(ctx)
		if err != nil {
			s.logger.Error(ctx, "error when begin transaction", logger.Error(err))
			return err
		}

		defer func() {
			if isTrxCompleted {
				return
			}
			errRollback := s.repo.RollbackTransaction(ctx)
			if errRollback != nil {
				s.logger.Error(ctx, "error when rollback transaction", logger.Error(errRollback))
			}
		}()
	}

	err = moneyFlowSvc.CreateTransactions(ctx, request)
	if err != nil {
		return err
	}
	if existingTrxCtx == nil {
		errCommit := s.repo.CommitTransaction(ctx)
		if errCommit != nil {
			s.logger.Error(ctx, "error when commit transaction", logger.Error(errCommit))
			return constant.ErrCommitDatabaseTransaction
		}
	}
	isTrxCompleted = true

	return nil
}
