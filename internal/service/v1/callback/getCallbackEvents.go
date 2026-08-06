package callbackService

import (
	"context"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *CallbackService) GetCallbackEvents(ctx context.Context) ([]callbackModel.CallbackEvent, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/GetCallbackEvents")
	defer segment.End()

	events, err := s.callbackRepo.GetCallbackEvents(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to get callback events", logger.Error(err))
		return nil, err
	}

	return events, nil
}
