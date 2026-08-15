package industry

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"
)

// Delete soft deletes an industry by setting deleted_at timestamp
func (r *repository) Delete(ctx context.Context, uuid string) error {
	ctx, span := otelTracer.Start(ctx, "IndustryRepository.Delete")
	defer span.End()

	now := time.Now().UTC()
	query := fmt.Sprintf(`
		UPDATE %s
		SET deleted_at = ?, updated_at = ?
		WHERE uuid = ? AND deleted_at IS NULL
	`, industriesTableName)

	_, err := r.db.ExecContext(ctx, query, now, now, uuid)
	if err != nil {
		r.logger.Error(ctx, "failed to soft delete industry", logger.Error(err), logger.String("uuid", uuid))
		return fmt.Errorf("failed to delete industry: %w", err)
	}

	return nil
}
