package bankAccount

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (b *bankAccountService) Update(
	ctx context.Context,
	request *bankAccount.UpdateBankAccountRequest) (*bankAccount.BankAccountResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/bankAccount/Update")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	// check if merchant has already registered bank account
	existed, err := b.repo.GetByMerchantID(ctx, request.MerchantID)
	if err != nil {
		b.logger.Error(ctx, "Error when getting bank account by merchant id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	if existed == nil {
		b.logger.Error(ctx, "Bank account not found", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrNotFound, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	bkAccount := &bankAccount.BankAccount{
		UUID:                   existed.UUID,
		MerchantID:             request.MerchantID,
		BeneficiaryAccountNo:   request.BeneficiaryAccountNo,
		BeneficiaryAccountName: request.BeneficiaryAccountName,
		BeneficiaryBankCode:    request.BeneficiaryBankCode,
		BeneficiaryBankName:    request.BeneficiaryBankName,
		CreatedBy:              existed.CreatedBy,
		CreatedAt:              existed.CreatedAt,
		UpdatedBy:              request.UpdatedBy,
		UpdatedAt:              time.Now(),
	}

	if err := b.repo.Update(ctx, bkAccount); err != nil {
		b.logger.Error(ctx, "Error when updating bank account", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	return &bankAccount.BankAccountResponse{
		BeneficiaryBankCode:    bkAccount.BeneficiaryBankCode,
		BeneficiaryBankName:    bkAccount.BeneficiaryBankName,
		BeneficiaryAccountNo:   bkAccount.BeneficiaryAccountNo,
		BeneficiaryAccountName: bkAccount.BeneficiaryAccountName,
	}, nil
}
