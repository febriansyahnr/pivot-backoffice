package adjustment

import (
	"context"

	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

const tableName = "manual_adjustment_histories"

func (r *adjustment) CreateAdjustment(ctx context.Context, data *adjustModel.ManualAdjustmentHistory) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/adjustment/CreateManualTopup")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		INSERT INTO ` + tableName + ` (
			uuid, merchant_id, transaction_date, bank_reference_id, bank_account, type, action,
			currency, amount, reference_id, proof_of_transfer, notes, created_by, created_at, updated_at
		) VALUES (
            :uuid, :merchant_id, :transaction_date, :bank_reference_id, :bank_account, :type, :action,
			:currency, :amount, :reference_id, :proof_of_transfer, :notes, :created_by, :created_at, :updated_at
        );`
	if _, err := r.db.NamedExecContext(ctx, query, data); err != nil {
		return err
	}
	return nil
}
