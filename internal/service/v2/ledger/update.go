package ledgerService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *LedgerService) UpdateTransaction(ctx context.Context, request *ledger_model.UpdateLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/ledger/RecordTransaction")
	defer segment.End()

	if err := request.Validate(); err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	err := s.repo.UpdateTransactionsStatus(ctx, request)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrUpdateTransactions)
	}

	return nil
}

func (s *LedgerService) BulkUpdateLedgerEntry(ctx context.Context, bulkRequest *ledger_model.BulkUpdateLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/ledger/BulkUpdateLedgerEntry")
	defer segment.End()

	if len(bulkRequest.Requests) == 0 {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload)
	}

	ctx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrInitDatabaseTransaction)
	}

	for _, request := range bulkRequest.Requests {
		request.ReferenceID = bulkRequest.ReferenceID
		err = s.UpdateTransaction(ctx, request)
		if err != nil {
			s.logger.Error(ctx, "error when update transaction", logger.Any("request", request))
			break
		}
	}
	if err != nil {
		errRollback := s.repo.RollbackTransaction(ctx)
		if errRollback != nil {
			s.logger.Error(ctx, "error when rollback transaction", logger.Error(errRollback))
			return pkgErrors.New(response.HttpErrInternal, constant.ErrRollbackDatabaseTransaction)
		}
		return err
	}

	err = s.repo.CommitTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "error commit transaction", logger.Error(err), logger.Any("request", bulkRequest))
		return pkgErrors.New(response.HttpErrInternal, constant.ErrCommitDatabaseTransaction)
	}
	return nil
}
