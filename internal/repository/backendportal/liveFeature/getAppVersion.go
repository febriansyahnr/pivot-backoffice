package liveFeature

import (
	"context"
	"encoding/json"
	"fmt"

	liveFeature "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/liveFeature"
)

func (r *LiveFeatureRepository) RetrieveAppVersion(ctx context.Context) (liveFeature.AppVersion, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/common/RetrieveAppVersion")
	defer segment.End()

	imageRetriever, err := r.consulRetrieverFactory(
		r.config.FeatureFlagConfig.ConsulAddr,
		r.config.FeatureFlagConfig.ConsulAppVersion,
		r.secret.ConsulSecret.Token,
	)
	if err != nil {
		return liveFeature.AppVersion{}, err
	}

	data, err := imageRetriever.Retrieve(ctx)
	if err != nil {
		return liveFeature.AppVersion{}, err
	}

	var response liveFeature.AppVersion
	if err = json.Unmarshal(data, &response); err != nil {
		return liveFeature.AppVersion{}, fmt.Errorf("failed to unmarshal data into AppVersion struct: %w", err)
	}

	return response, nil
}
