package withdrawalService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *withdrawalService) GetList(ctx context.Context, request *withdrawal.WithdrawalHistoryRequest) (list *commonModel.PaginationResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/GetList")
	defer segment.End()

	if list, err = s.repo.GetList(ctx, request); err != nil {
		s.logger.Error(ctx, "Get list withdrawal history", logger.Error(err))
	}
	return
}

func (s *withdrawalService) GetById(ctx context.Context, request *withdrawal.WithdrawalDetailRequest) (result *withdrawal.WithdrawalDetailResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/GetById")
	defer segment.End()

	if result, err = s.repo.GetById(ctx, request); err != nil {
		s.logger.Error(ctx, "Get withdrawal details", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if result == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}
	return
}

// GetTodayWithdrawalInsight return total withdrawal and the amount based on the withdrawal status
// it will calculate the data only for today
// when it found error, then it will return the error with nil withdrawal item
func (s *withdrawalService) GetTodayWithdrawalInsight(ctx context.Context, opt withdrawal.WithdrawalInsightRequest) (*withdrawal.WithdrawalInsightResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/GetTodayWithdrawalInsight")
	defer segment.End()

	return s.repo.GetTodayWithdrawalInsight(ctx, opt)
}
