package reconciliation

import (
	"context"
	"database/sql"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *ReconciliationService) ProcessPayoutRecon(ctx context.Context, req *reconciliation.ReconciliationPayout) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/ProcessPayoutRecon")
	defer segment.End()

	s.logger.Info(ctx, "ProcessPayoutRecon - Processing payout reconciliation", logger.Any("request", req))

	// check existing account_transactions
	accountTrx, err := s.accountTransactionRepo.FindByID(
		ctx,
		req.ExternalID,
	)
	if err != nil {
		s.logger.Error(ctx, "ProcessPayoutRecon - error when get transaction by processor id", logger.Error(err))
		if err == sql.ErrNoRows {
			return nil
		}

		return err
	}

	if accountTrx == nil {
		s.logger.Info(ctx, "ProcessPayoutRecon - Transaction not found", logger.Any("external_id", req.ExternalID))
		return nil
	}

	amountValue := "0"
	if req.Amount != nil {
		amountValue = req.Amount.Value
	}

	// update account_transactions.AdditionalInfo.ReconDetail
	if err := s.accountTransactionRepo.SetAdditionalInfoReconciliation(
		ctx,
		accountTrx.UUID.String(),
		&reconciliation.ReconDetail{
			Status:   req.Status,
			Reason:   req.Reason,
			Amount:   amountValue,
			DateTime: time.Now().UTC().String(),
		},
	); err != nil {
		s.logger.Error(ctx, "ProcessPayoutRecon - error when update transaction by processor id", logger.Error(err))
		return err
	}

	s.logger.Info(ctx, "ProcessPayoutRecon - Transaction updated successfully")

	return nil
}
