package payoutManualProcessingAccount

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PayoutManualProcessingAccountService) List(
	ctx context.Context,
	req *payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery,
) (*commonModel.PaginationResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/payoutManualProcessingAccount/List")
	defer span.End()

	list, total, err := s.repo.List(ctx, req)
	if err != nil {
		s.logger.Error(ctx, "error when listing payout manual processing accounts", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrGetPayoutManualProcessingAccountList)
	}

	return &commonModel.PaginationResponse{
		Data: list,
		Meta: *commonModel.NewMeta(req.Page, req.PageSize, int64(total)),
	}, nil
}
