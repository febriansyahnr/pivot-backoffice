package statusHistoriesRepository

import (
	"context"

	statusHistoriesModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/statusHistory"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

const tableName = "status_histories"

func (r *statusHistoriesRepo) Insert(ctx context.Context, data *statusHistoriesModel.StatusHistory) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/statusHistories/Insert")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		INSERT INTO ` + tableName + ` (
			id, reference_type, reference_id, status, metadata, created_at, updated_at
		) VALUES (
			:id, :reference_type, :reference_id, :status, :metadata, :created_at, :updated_at
		)`

	if _, err := r.db.NamedExecContext(ctx, query, data); err != nil {
		return err
	}

	return nil
}
