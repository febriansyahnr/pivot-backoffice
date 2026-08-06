package merchantTopUp

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/topUpSimulation"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *merchantTopUpService) CreateTopupSimulation(ctx context.Context, req snapCoreModel.TopupSimulationRequest) (*snapCoreModel.TopupSimulationResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantTopUp/CreateTopupSimulation")
	defer segment.End()

	if s.config.Environment == constant.EnvironmentProduction {
		return nil, pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenAccess)
	}

	topUpRef, err := s.merchantTopUpRepo.GetByReferenceNumber(ctx, req.VANumber)
	if err != nil {
		s.logger.Error(ctx, "Failed while get merchant top up reference by reference number", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if topUpRef == nil || topUpRef.AccountName != req.AccountName {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)

	} else if topUpRef.MerchantID != req.MerchantId {
		return nil, pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenAccess)
	}

	res, err := s.snapCore.TopUpSimulation(ctx, req)
	if err != nil {
		s.logger.Error(ctx, "Failed to create top up simulation VA", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, err)
	}
	return res, nil
}
