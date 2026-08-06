package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) CloseMerchant(ctx context.Context, merchant *merchantModel.UpdateMerchantStatus) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/CloseMerchant")
	defer segment.End()

	// check if merchant is exist
	merchantExist, err := s.repo.FindMerchantByID(ctx, merchant.ID)
	if err != nil {
		s.logger.Error(ctx, "error when find merchant by id", logger.Error(err))
		return pkgErrors.New(responseHttp.HttpErrInternal, err)
	} else if merchantExist == nil {
		return pkgErrors.New(responseHttp.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	availableBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, merchant.ID, constant.TypeDisbursement)
	if err != nil {
		s.logger.Error(ctx, "error when get merchant balance", logger.Error(err), logger.Any("merchantID", merchant.ID))
		return constant.ErrValidateBalance

	}

	if availableBalance != 0 {
		s.logger.Error(ctx, "error when update merchant status", logger.Error(err), logger.Any("merchantID", merchant.ID))
		return constant.ErrMerchantNotAllowedUpdateStatus
	}

	if errUpdate := s.UpdateStatusByID(ctx,
		merchant.Status,
		merchant.ReasonStatus,
		merchant.ID,
	); errUpdate != nil {
		return errUpdate
	}

	return nil
}
