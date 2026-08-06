package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) FindMerchantFeeByMerchantIDAndType(
	ctx context.Context, merchantId, reference string) (*merchantModel.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/FindMerchantFeeByMerchantIDAndType")
	defer segment.End()

	merchantFee, err := s.repo.GetMerchantFeeByMerchantIDAndType(ctx, merchantId, reference)
	if err != nil {
		s.logger.Error(ctx, "error when finding merchant by merchant id and reference", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrInternal, err)
	}

	if merchantFee == nil {
		return nil, errors.New(responseHttp.HttpErrNotFound, constant.ErrDataNotFound)
	}

	return merchantFee, nil
}
