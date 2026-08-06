package vendor

import (
	"context"
	"time"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *VendorRepository) Delete(ctx context.Context, id string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/vendor/Delete")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	now := time.Now().UTC()
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE `+tableName+` SET deleted_at = ?, updated_at = ? WHERE uuid = ?;`, now, now, id,
	)
	if err != nil {
		r.logger.Error(ctx, "error when deleting vendor", logger.Error(err), logger.String("id", id))
	}
	return err
}
