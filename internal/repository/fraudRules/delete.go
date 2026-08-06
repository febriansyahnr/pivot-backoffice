package fraudrulesrepository

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *FraudRulesRepository) Delete(ctx context.Context, id string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/fraudRules/Delete")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	now := time.Now().UTC()
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE `+tableName+` SET deleted_at = ?, updated_at = ? WHERE uuid = ?;`, now, now, id,
	)
	if err != nil {
		r.logger.Error(ctx, "error when deleting fraud rule", logger.Error(err), logger.Any("id", id))
	}
	return err
}
