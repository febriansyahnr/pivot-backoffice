package reconciliation

import (
	"context"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Create implements repository.IReconciliationRepository.
func (r *ReconciliationRepository) Create(ctx context.Context, data *reconciliation.Reconciliation) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/reconciliation/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := fmt.Sprintf(`
		INSERT INTO %s (uuid, original_name, file_path, result_file_path, status, transaction_type, reasons, created_by, created_at, updated_at)
		VALUES (:uuid, :original_name, :file_path, :result_file_path, :status, :transaction_type, :reasons, :created_by, :created_at, :updated_at)
	`, tableName)

	affected, err := r.db.NamedExecContext(ctx, query, data)
	if err != nil {
		r.logger.Error(ctx, "error when inserting reconciliation", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when inserting reconciliation", logger.Error(errors.New("no rows affected")))
		return constant.ErrNoRowsAffected
	}

	return nil
}
