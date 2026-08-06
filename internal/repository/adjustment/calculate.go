package adjustment

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *adjustment) CalculateAmountBalanceAdjustmentForTopUp(ctx context.Context, relatedAdjustmentID string) (float64, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/adjustment/CalculateAmountBalanceAdjustmentForTopUp")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	sumAmount := 0.0

	query := `SELECT COALESCE(SUM(amount),0) FROM ` + tableName + ` WHERE uuid = ? OR reference_id = ?`
	err := r.db.GetContext(ctx, &sumAmount, query, relatedAdjustmentID, relatedAdjustmentID)
	if err != nil {
		return sumAmount, err
	}

	return sumAmount, nil
}
