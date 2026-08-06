package payoutManualProcessingAccount

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PayoutManualProcessingAccountService) Create(
	ctx context.Context,
	request *payoutManualProcessingAccountModel.CreatePayoutManualProcessingAccountRequest,
) (*payoutManualProcessingAccountModel.PayoutManualProcessingAccountResponse, error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/payoutManualProcessingAccount/Create")
	defer span.End()

	account := request.ToPayoutManualProcessingAccount()

	err := s.repo.Create(ctx, account)
	if err != nil {
		if strings.Contains(err.Error(), "1062") && strings.Contains(err.Error(), "Duplicate entry") {
			return nil, pkgErrs.New(response.HttpErrRequest, constant.ErrPayoutManualProcessingAccountAlreadyExists)
		}
		s.logger.Error(ctx, "error when creating payout manual processing account", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}

	return account.ToResponse(), nil
}
