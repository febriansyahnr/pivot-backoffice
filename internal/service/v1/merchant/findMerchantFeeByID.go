package merchant

import (
	"context"
	"fmt"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) FindMerchantFeeByID(ctx context.Context, id string) (*merchantModel.MerchantFee, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/FindMerchantFeeByID")
	defer segment.End()

	merchantFee, err := s.repo.GetMerchantFeeByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "error when finding merchant by id", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrInternal, err)
	}

	if merchantFee == nil {
		return nil, errors.New(responseHttp.HttpErrNotFound, fmt.Errorf("merchant fee not found"))
	}

	return merchantFee, nil
}
