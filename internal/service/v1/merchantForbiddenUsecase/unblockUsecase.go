package merchantForbiddenUsecase

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantforbiddenusecaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantForbiddenUsecase"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantForbiddenUseCaseService) UnblockUseCase(ctx context.Context,
	request *merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest) error {

	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantforbiddenusecase/UnblockMerchantUseCase")
	defer segment.End()

	if request.Requester != constant.UserSystemType {
		err := s.merchantService.ValidateSubMerchantParent(ctx, request.Requester, request.MerchantID)
		if err != nil {
			return pkgErrs.New(response.HttpErrForbidden, constant.ErrIncorrectSubMerchant)
		}
	}

	merchant, err := s.merchantService.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return err
	}
	if merchant == nil {
		return pkgErrs.New(response.HttpErrRequest, constant.ErrIncorrectMerchantID)
	}

	request.UseCase = strings.ToUpper(request.UseCase)
	forbiddenList, err := s.repo.GetForbiddenUsecase(ctx, &merchantforbiddenusecaseModel.GetMerchantForbiddenUseCaseRequest{
		MerchantID: request.MerchantID,
		UseCase:    request.UseCase,
	})
	if err != nil || forbiddenList == nil {
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrBlockedUnblockedMerchantUseCase)
	}
	if len(forbiddenList) == 0 {
		return nil
	}

	err = s.repo.RemoveForbiddenUsecase(ctx, forbiddenList[0])
	if err != nil {
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrBlockedUnblockedMerchantUseCase)
	}

	if request.SetStatus {
		err = s.merchantService.UpdateStatusByID(ctx, constant.MerchantStatusActive, "reactivated by parent merchant via dashboard", request.MerchantID)
		if err != nil {
			return pkgErrs.New(response.HttpErrDatabase, constant.ErrBlockedUnblockedMerchantUseCase)
		}
	}

	err = s.rabbitMqExt.PublishActivity(ctx, &request.MerchantID, &request.Requester, constant.TagMerchantForbiddenUseCase, constant.ActivityBlockMerchantDisbursement, request)
	if err != nil {
		s.logger.Error(ctx, "unable to publish merchant blocked/unblocked use case activity.", logger.Error(err), logger.Any("request", request))
	}

	return nil
}
