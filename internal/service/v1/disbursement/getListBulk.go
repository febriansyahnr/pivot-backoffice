package disbursementService

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
)

func (s *DisbursementService) GetListBulk(ctx context.Context, filter *disbursementModel.GetBulkDisbursementFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetListBulk")
	defer segment.End()

	list, err := s.disbursementRepo.GetListBulk(ctx, filter, page, perPage)
	if err != nil {
		return nil, err
	}

	return list, nil
}
