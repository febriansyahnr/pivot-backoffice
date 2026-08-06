package merchantTopUp

import (
	"context"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
)

func (s *merchantTopUpService) GetList(ctx context.Context, request *model.TopUpTransactionListRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantTopUp/GetList")
	defer segment.End()

	list, err := s.merchantTopUpRepo.GetList(ctx, request)
	if err != nil {
		return nil, err
	}

	return list, nil
}
