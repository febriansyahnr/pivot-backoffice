package bankAccount

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (svc *bankAccountService) GetByMerchantID(ctx context.Context, request *bankAccount.GetMerchantBankAccountRequest) (*bankAccount.BankAccount, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/bankAccount/GetByMerchantID")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	bankAccount, err := svc.repo.GetByMerchantID(ctx, request.MerchantID)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	if bankAccount == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, constant.ErrBankAccountNotFound)
	}

	return bankAccount, nil
}
