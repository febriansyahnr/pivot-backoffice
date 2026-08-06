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

func (s *MerchantForbiddenUseCaseService) BlockUseCase(ctx context.Context,
	request *merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest) error {

	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantforbiddenusecase/BlockMerchantUseCase")
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
	if !merchantforbiddenusecaseModel.IsUseCaseExists(request.UseCase) {
		return pkgErrs.New(response.HttpErrRequest, constant.ErrForbiddenUseCaseNotExist)
	}

	data, err := s.repo.GetForbiddenUsecase(ctx, &merchantforbiddenusecaseModel.GetMerchantForbiddenUseCaseRequest{
		MerchantID: request.MerchantID,
		UseCase:    request.UseCase,
	})
	if err != nil {
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrBlockedUnblockedMerchantUseCase)
	}

	if len(data) == 0 {
		newData := merchantforbiddenusecaseModel.NewMerchantForbiddenUseCase(&merchantforbiddenusecaseModel.NewMerchantForbiddenUseCaseRequest{
			MerchantID: request.MerchantID,
			UseCase:    request.UseCase,
		})
		_, err := s.repo.RegisterForbiddenUsecase(ctx, newData)
		if err != nil {
			return pkgErrs.New(response.HttpErrDatabase, constant.ErrBlockedUnblockedMerchantUseCase)
		}

		if request.SetStatus {
			err = s.merchantService.UpdateStatusByID(ctx, constant.MerchantStatusDeactivated, "deactivated by parent merchant via dashboard", request.MerchantID)
			if err != nil {
				return pkgErrs.New(response.HttpErrDatabase, constant.ErrBlockedUnblockedMerchantUseCase)
			}
		}

	}

	err = s.rabbitMqExt.PublishActivity(ctx, &request.MerchantID, &request.Requester, constant.TagMerchantForbiddenUseCase, constant.ActivityBlockMerchantDisbursement, request)
	if err != nil {
		s.logger.Error(ctx, "unable to publish merchant blocked/unblocked use case activity.", logger.Error(err), logger.Any("request", request))
	}

	return nil
}
