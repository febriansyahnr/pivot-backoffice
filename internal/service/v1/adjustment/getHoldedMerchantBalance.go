package adjustment

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *adjustmentService) GetHoldedMerchantBalance(ctx context.Context, req *adjustModel.GetHoldedMerchantBalanceRequest) (*adjustModel.GetHoldedMerchantBalanceResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/adjustment/GetHoldedMerchantBalance")
	defer segment.End()

	merchantID, _ := uuid.Parse(req.MerchantId)
	account, err := s.accountRepo.FindMerchantAccountByName(ctx, merchantID, req.AccountType)
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrDatabase, err)
	}
	if account == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrAccountNotFound)
	}

	return &adjustModel.GetHoldedMerchantBalanceResponse{
		Amount:      account.HoldedBalance,
		MerchantID:  req.MerchantId,
		AccountType: req.AccountType,
	}, nil
}
