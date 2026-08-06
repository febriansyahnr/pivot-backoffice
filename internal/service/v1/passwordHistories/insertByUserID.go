package passwordHistories

import (
	"context"

	"github.com/google/uuid"
)

func (s *PasswordHistoriesService) InsertByUserID(ctx context.Context, userID string, password string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/passwordHistories/InsertByUserID")
	defer segment.End()

	passwordHistoryId := uuid.New().String()
	return s.repo.Insert(ctx, passwordHistoryId, userID, password)
}
