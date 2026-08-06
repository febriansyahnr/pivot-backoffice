package role

import (
	"context"

	roleModel "github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RoleRepository) Update(ctx context.Context, role *roleModel.Role) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/role/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "roles")
	query := `
			UPDATE
				roles
			SET slug = ?, merchant_id = ?, name = ?, type = ?, created_at = ?, updated_at = ?, deleted_at = ?
			WHERE
			    uuid = ?`
	_, err := r.db.ExecContext(
		ctx, query,
		role.Slug,
		role.MerchantID,
		role.Name,
		role.Type,
		role.CreatedAt,
		role.UpdatedAt,
		role.DeletedAt,
		role.UUID,
	)
	if err != nil {
		r.logger.Error(ctx, "error when updating role", logger.Error(err))
		return err
	}

	return nil
}
