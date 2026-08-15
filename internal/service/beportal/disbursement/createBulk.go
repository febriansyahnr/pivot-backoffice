package disbursementService

import (
	"context"
	"time"

	"github.com/google/uuid"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
)

func (s *DisbursementService) CreateBulk(ctx context.Context, request *disbursementModel.CreateBulkDisbursementRequest) (*disbursementModel.BulkDisbursement, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/CreateBulk")
	defer segment.End()

	bulkDisbursement := &disbursementModel.BulkDisbursement{
		UUID:       uuid.NewString(),
		MerchantID: request.MerchantID,
		File:       request.File,
		Status:     request.Status,
		CreatedBy:  &request.CreatedBy,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	err := s.disbursementRepo.InsertBulkDisbursement(ctx, bulkDisbursement)
	if err != nil {
		return nil, err
	}

	return bulkDisbursement, nil
}
