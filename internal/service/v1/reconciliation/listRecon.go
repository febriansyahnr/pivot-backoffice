package reconciliation

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// ListRecon returns list of recon
func (r *ReconciliationService) ListRecon(ctx context.Context, request *reconciliation.ReconciliationFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/ListRecon")
	defer segment.End()

	recons, err := r.reconRepo.GetAll(ctx, request)
	if err != nil {
		r.logger.Error(ctx, "Failed to get list recon", logger.Error(err))
		return &commonModel.PaginationResponse{
			Data: []map[string]string{},
		}, err
	}

	if reconList, ok := recons.Data.([]*reconciliation.Reconciliation); ok {
		responseData := make([]*reconciliation.ReconciliationResponse, len(reconList))
		for i, recon := range reconList {
			responseData[i] = recon.ToResponse()
		}
		recons.Data = responseData
	}

	return recons, nil
}
