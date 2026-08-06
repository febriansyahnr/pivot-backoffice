package tnc

import (
	"context"
	"time"

	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *TNCRepository) UpdateTNCVersion(ctx context.Context, version *tncModel.TNC) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/UpdateTNCVersion")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, versionsTableName)

	query := `
		UPDATE tncs SET
			version = :version,
			title = :title,
			markdown_content = :markdown_content,
			is_active = :is_active,
			updated_at = :updated_at
		WHERE uuid = :uuid AND deleted_at IS NULL`

	version.UpdatedAt = time.Now().UTC()
	_, err := r.db.NamedExecContext(ctx, query, version)
	if err != nil {
		r.logger.Error(ctx, "error when updating tnc version", logger.Error(err), logger.String("id", version.UUID))
		return err
	}

	return nil
}

func (r *TNCRepository) DeactivateAllTNCVersions(ctx context.Context) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/tnc/DeactivateAllTNCVersions")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, versionsTableName)

	query := `UPDATE tncs SET is_active = 0, updated_at = ? WHERE is_active = 1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, time.Now().UTC())
	if err != nil {
		r.logger.Error(ctx, "error when deactivating all tnc versions", logger.Error(err))
		return err
	}

	return nil
}
