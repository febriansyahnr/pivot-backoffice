package callbackService

import (
	"context"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

func (s *CallbackService) GetList(ctx context.Context, filter *callbackModel.GetListCallbackFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/GetList")
	defer segment.End()

	list, err := s.callbackRepo.GetList(ctx, filter, page, perPage)
	if err != nil {
		return nil, err
	}

	return list, nil
}
