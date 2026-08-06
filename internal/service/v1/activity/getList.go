package activityService

import (
	"context"

	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

func (s *ActivityService) GetList(
	ctx context.Context,
	filter activityModel.ActivityFilterRequest,
	page, perPage int64,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/activity/GetList")
	defer segment.End()

	list, err := s.repo.GetList(ctx, filter, page, perPage)
	if err != nil {
		return nil, err
	}

	return list, nil
}
