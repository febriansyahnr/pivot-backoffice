package reconciliation

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
)

// UpdateReconDetail implements service.IReconciliationService.
func (s *ReconciliationService) UpdateReconDetail(ctx context.Context, id string, payload *reconciliation.ReconDetail) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/UpdateReconDetail")
	defer segment.End()

	// Find account transaction by id
	accountTransaction, err := s.accountTransactionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if accountTransaction == nil {
		return errors.New("account transaction not found")
	}

	// update recon  detail
	if err := s.accountTransactionRepo.UpdateReconDetail(ctx, id, payload); err != nil {
		return err
	}

	return nil
}
