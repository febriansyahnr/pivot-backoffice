package disbursementService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
)

func (s *DisbursementService) ValidateBalance(ctx context.Context, request *disbursementModel.ValidateBalanceRequest) bool {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ValidateBalance")
	defer segment.End()

	availableBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, request.MerchantID, constant.TypeDisbursement)
	if err != nil {
		return false
	}

	sumAmountObject, err := s.disbursementRepo.SumAmountByIDs(ctx, request.DisbursementIDs)
	if err != nil {
		return false
	}

	if sumAmountObject.TotalAmount > availableBalance {
		return false
	}

	// Transactions on Behalf
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {

		parentBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, parentMerchantId, constant.TypeDisbursement)
		if err != nil {
			return false

		} else if sumAmountObject.ParentFeeCharged > parentBalance {
			return false
		}
	}
	return true
}
