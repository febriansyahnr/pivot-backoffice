package role

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *RoleRepository) Delete(ctx context.Context, id string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/role/Delete")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "roles")

	now := time.Now().UTC()
	_, err = r.db.ExecContext(
		ctx,
		`UPDATE roles SET deleted_at = ?, updated_at = ? WHERE uuid = ?;`, now, now, id,
	)
	return
}
