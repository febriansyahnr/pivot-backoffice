package disbursementService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) GetBulkDisbursementDetail(ctx context.Context, id string) (*disbursementModel.BulkDisbursementDetail, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetBulkDisbursementDetail")
	defer segment.End()

	detail, err := s.disbursementRepo.GetBulkDisbursementDetailByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "error when finding bulk disbursement detail by id", logger.String("uuid", id), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if detail == nil {
		s.logger.Warn(ctx, "bulk disbursement not found", logger.String("uuid", id))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrBulkDisbursementNotFound)
	}

	return detail, nil
}
