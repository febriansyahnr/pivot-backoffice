package merchant

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) ListSubMerchantByParentID(
	ctx context.Context,
	filter *merchant.SubMerchantListFilter,
	page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/ListSubMerchantByParentID")
	defer segment.End()

	subMerchants, err := s.repo.ListSubMerchantByParentID(ctx, filter, page, perPage)
	if err != nil {
		s.logger.Error(ctx, "Failed to get list sub merchants", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	return subMerchants, nil
}
