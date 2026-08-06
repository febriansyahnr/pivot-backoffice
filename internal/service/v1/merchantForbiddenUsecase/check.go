package merchantForbiddenUsecase

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantforbiddenusecaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantForbiddenUsecase"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantForbiddenUseCaseService) CheckUseCase(ctx context.Context, merchantId, useCase string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantforbiddenusecase/CheckUseCase")
	defer segment.End()

	if !merchantforbiddenusecaseModel.IsUseCaseExists(useCase) {
		return nil
	}

	forbidList, err := s.repo.GetForbiddenUsecase(ctx, &merchantforbiddenusecaseModel.GetMerchantForbiddenUseCaseRequest{
		MerchantID: merchantId,
		UseCase:    strings.ToUpper(useCase),
	})
	if err != nil {
		s.logger.Error(ctx, "error when validate merchant use case", logger.Error(err), logger.Any("merchantId", merchantId), logger.Any("useCase", useCase))
		return constant.ErrFailedValidateUseCase
	}

	if len(forbidList) > 0 {
		return errors.New(response.HttpErrForbidden, constant.ErrForbiddenUseCase)
	}
	return nil
}
