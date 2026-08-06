package installmentplan

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *InstallmentPlanService) List(ctx context.Context, req *installmentPlanModel.FilterRequest) ([]*installmentPlanModel.InstallmentPlan, int64, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/installmentPlan/List")
	defer span.End()

	list, count, err := s.repo.List(ctx, req)
	if err != nil {
		s.logger.Error(ctx, "failed to list installment plan", logger.Error(err))
		return nil, 0, pkgErrors.New(response.HttpErrDatabase, constant.ErrGetInstallmentPlan)
	}

	return list, count, nil
}
