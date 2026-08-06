package liveFeature

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	liveFeatureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/liveFeature"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func (s *LiveFeatureService) GetAppVersion(ctx context.Context) (liveFeatureModel.AppVersion, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/liveFeature/GetAppVersion")
	defer segment.End()

	data, err := s.repo.RetrieveAppVersion(ctx)
	if err != nil {
		return liveFeatureModel.AppVersion{}, fmt.Errorf("failed to retrieve app version: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, newVersion := range data.Versions {
		oldVersion, exists := s.currentVersion.Versions[key]
		if !exists || oldVersion != newVersion {
			if err := s.notifyVersionChange(ctx, key, newVersion); err != nil {
				s.logger.Error(ctx, "failed to send notification", logger.String("key", key), logger.Error(err))
			}
		}
	}

	s.currentVersion = data

	return data, nil
}

func (s *LiveFeatureService) notifyVersionChange(ctx context.Context, key, version string) error {
	notificationPayload := notification.PushNotification{
		RoutingKey: fmt.Sprintf(constant.NotificationRoutingKeyFmt, key),
		Payload: notification.PushNotificationPayload{
			ID:        uuid.NewString(),
			Subject:   "Please reload your browser!",
			Type:      version,
			Message:   fmt.Sprintf("The %s application has been updated to version <b>%s</b>.", key, version),
			CreatedAt: time.Now().UTC(),
		},
	}

	// Push notification to RabbitMQ
	return s.rabbitMqExt.PushNotification(ctx, &notificationPayload)
}

func (s *LiveFeatureService) PollForChanges(ctx context.Context, interval time.Duration, config *config.Config) {

	ffContext := ffcontext.NewEvaluationContext(config.Environment)
	ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, config.Environment)
	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagKeyEnableFrontEndAppVersion, ffContext, false)

	if !enabled {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := s.GetAppVersion(ctx); err != nil {
				s.logger.Error(ctx, "Error retrieving app version", logger.Error(err))
			}
		}
	}
}
