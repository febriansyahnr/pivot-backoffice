package ipWhitelistRepository

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *IPWhitelistRepository) Delete(ctx context.Context, id string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/ipWhitelist/Delete")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, IPWhitelistTable)

	now := time.Now().UTC()
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE `+IPWhitelistTable+` SET deleted_at = ?, updated_at = ? WHERE id = ?;`, now, now, id,
	)
	if err != nil {
		r.logger.Error(ctx, "error when deleting ip whitelist", logger.Error(err), logger.Any("id", id))
	}
	return err
}
