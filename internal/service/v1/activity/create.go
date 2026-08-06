package activityService

import (
	"context"

	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
)

func (s *ActivityService) Create(ctx context.Context, activity *activityModel.Activity) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/activity/Create")
	defer segment.End()

	return s.repo.Create(ctx, activity)
}
