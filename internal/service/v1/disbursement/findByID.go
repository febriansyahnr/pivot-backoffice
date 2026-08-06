package disbursementService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) FindByID(ctx context.Context, id string) (*disbursementModel.DisbursementWithTransaction, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/FindByID")
	defer segment.End()

	disbursement, err := s.disbursementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if disbursement == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrDisbursementNotFound)
	}
	referenceType := constant.TypeDisbursement
	if disbursement.MetadataObj.XbDetail != nil && disbursement.MetadataObj.XbDetail.Uuid != "" {
		referenceType = constant.TypeXB
	}

	// Build status history
	statusHistories, err := s.statusHistoriesRepo.GetByReference(ctx, referenceType, id)
	if err != nil {
		s.logger.Error(ctx, "error when get status history by reference", pdkLogger.String("disbursementID", id))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if len(statusHistories) > 0 {
		disbursement.StatusHistories = statusHistories
	}

	return disbursement, err
}
