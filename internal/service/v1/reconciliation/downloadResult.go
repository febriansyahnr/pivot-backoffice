package reconciliation

import (
	"context"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *ReconciliationService) DownloadResult(ctx context.Context, id string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/reconciliation/DownloadResult")
	defer segment.End()

	reconDetail, err := s.reconRepo.GetByUUID(ctx, id)
	if err != nil {
		return "", err
	}
	url, err := s.gcs.CreateSignedURL(ctx, reconDetail.ResultFilePath, 5*time.Minute)
	if err != nil {
		s.logger.Error(ctx, "Failed to get signed url from gcs", logger.Error(err))
		return "", err
	}

	return url, nil
}
