package adjustment

import (
	"context"
	"database/sql"
	"errors"

	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *adjustment) FindByID(ctx context.Context, id string) (*adjustModel.ManualAdjustmentHistory, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/adjustment/FindByID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var adjustmentHistory adjustModel.ManualAdjustmentHistory
	query := `
			SELECT 
				uuid, merchant_id, transaction_date, bank_reference_id, bank_account, type, action,
				currency, amount, reference_id, proof_of_transfer, notes, created_by, created_at, updated_at
			FROM ` + tableName + ` 
			WHERE uuid = ?`

	if err := r.db.GetContext(ctx, &adjustmentHistory, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &adjustmentHistory, nil
}
