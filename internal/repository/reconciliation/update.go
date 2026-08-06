package reconciliation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *ReconciliationRepository) Update(ctx context.Context, data *reconciliation.Reconciliation) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/reconciliation/Update")
	defer segment.End()

	data.UpdatedAt = time.Now().UTC()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	query := fmt.Sprintf(`
		UPDATE %s
		SET
			updated_at = :updated_at,
			result_file_path = :result_file_path,
			reasons = :reasons,
			status = :status
		WHERE uuid = :uuid
	`,
		tableName,
	)

	_, err := r.db.NamedExecContext(
		ctx,
		query,
		data,
	)

	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating reconciliation",
			logger.Error(err),
			logger.Any("request", data),
			logger.String("query", query))
		return err
	}

	return nil
}
