package payoutManualProcessingAccount

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PayoutManualProcessingAccountService) Update(
	ctx context.Context,
	request *payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest,
) (*payoutManualProcessingAccountModel.PayoutManualProcessingAccountResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/payoutManualProcessingAccount/Update")
	defer span.End()

	account, err := s.repo.GetByUUID(ctx, request.UUID)
	if err != nil {
		s.logger.Error(ctx, "error when getting payout manual processing account for update", logger.Error(err), logger.Any("uuid", request.UUID))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	if account == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrPayoutManualProcessingAccountNotFound)
	}

	if request.Status == nil || (*request.Status != constant.StatusActive && *request.Status != constant.StatusInactive) {
		return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidPayoutManualProcessingAccountStatus)
	}

	account.Update(request)

	err = s.repo.Update(ctx, account)
	if err != nil {
		s.logger.Error(ctx, "error when updating payout manual processing account", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	return account.ToResponse(), nil
}
