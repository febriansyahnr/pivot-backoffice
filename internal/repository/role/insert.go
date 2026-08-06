package role

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RoleRepository) Create(ctx context.Context, roles *role.Role) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/role/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "roles")

	query := `
		INSERT INTO roles (uuid, merchant_id, name, slug, type, created_at, updated_at, deleted_at)
		VALUES (:uuid, :merchant_id, :name, :slug, :type, :created_at, :updated_at, :deleted_at)
	`

	affected, err := r.db.NamedExecContext(ctx, query, roles)
	if err != nil {
		r.logger.Error(ctx, "error when inserting role", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when inserting role", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}

	return nil
}
