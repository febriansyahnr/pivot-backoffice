package liveFeature

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/liveFeature"
)

func (s *LiveFeatureService) GetList(ctx context.Context) ([]liveFeature.LiveFeature, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/liveFeature/GetList")
	defer segment.End()

	list, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return list, nil
}
