package disbursementService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) GetMerchantTransactionConfig(ctx context.Context, merchantId string) (*disbursementModel.TransactionConfig, context.Context, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetMerchantTransactionConfig")
	defer segment.End()

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "Failed while find merchant by id", logger.Error(err))
		return nil, nil, pkgErrs.New(response.HttpErrDatabase, err)

	} else if merchant == nil {
		return nil, nil, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantNotFound)
	}

	merchantConfigID := merchantId
	if merchant.ParentID.String != "" { // Sub-Merchant
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)

		if merchant.KYCStatus.String == constant.KYCStatusNotRequired {
			merchantConfigID = merchant.ParentID.String
		}
	}
	config, err := s.GetTransactionConfig(ctx, merchantConfigID)
	return config, ctx, err
}
