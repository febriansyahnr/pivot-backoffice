package beneficiaryAccountService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

func (s *BeneficiaryAccountService) GetList(
	ctx context.Context,
	filter *beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest,
	page, perPage int64,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/BeneficiaryAccountService/GetList")
	defer segment.End()

	if derivedID, ok := ctx.Value(constant.CtxDerivedMerchantID).(string); ok && derivedID != "" {
		list, err := s.beneficiaryAccountRepo.GetListOfDerived(ctx, filter, page, perPage)
		if err != nil {
			return nil, err
		}

		return list, nil
	}

	list, err := s.beneficiaryAccountRepo.GetList(ctx, filter, page, perPage)
	if err != nil {
		return nil, err
	}

	return list, nil
}
