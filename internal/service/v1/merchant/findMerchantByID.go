package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) FindMerchantByID(ctx context.Context, id string) (*merchantModel.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/FindMerchantByID")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "error when finding merchant by id", logger.Error(err), logger.String("id", id))
		return nil, errors.New(responseHttp.HttpErrInternal, constant.ErrFindMerchant)
	}

	return merchant, nil
}
