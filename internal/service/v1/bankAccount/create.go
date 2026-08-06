package bankAccount

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (b *bankAccountService) Create(
	ctx context.Context,
	request *bankAccount.CreateBankAccountRequest) (*bankAccount.BankAccountResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/bankAccount/Create")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	// check if merchant has already registered bank account
	existed, err := b.repo.GetByMerchantID(ctx, request.MerchantID)
	if err != nil {
		b.logger.Error(ctx, "Error when getting bank account by merchant id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	if existed != nil {
		b.logger.Error(ctx, "Bank account already exist", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrConflict, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	bkId, _ := uuid.NewV7()
	bkAccount := &bankAccount.BankAccount{
		UUID:                   bkId.String(),
		MerchantID:             request.MerchantID,
		BeneficiaryAccountNo:   request.BeneficiaryAccountNo,
		BeneficiaryAccountName: request.BeneficiaryAccountName,
		BeneficiaryBankCode:    request.BeneficiaryBankCode,
		BeneficiaryBankName:    request.BeneficiaryBankName,
		CreatedBy:              request.CreatedBy,
		CreatedAt:              time.Now(),
		UpdatedBy:              request.CreatedBy,
		UpdatedAt:              time.Now(),
	}

	if err := b.repo.Create(ctx, bkAccount); err != nil {
		b.logger.Error(ctx, "Error when creating bank account", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	return &bankAccount.BankAccountResponse{
		BeneficiaryBankCode:    bkAccount.BeneficiaryBankCode,
		BeneficiaryBankName:    bkAccount.BeneficiaryBankName,
		BeneficiaryAccountNo:   bkAccount.BeneficiaryAccountNo,
		BeneficiaryAccountName: bkAccount.BeneficiaryAccountName,
	}, nil
}
