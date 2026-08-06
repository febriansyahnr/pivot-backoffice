package installmentPlan

import (
	"context"

	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *InstallmentPlanRepository) Update(ctx context.Context, plan *installmentPlanModel.InstallmentPlan) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/installmentPlan/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, installmentPlansTableName)
	query := `
		UPDATE ` + installmentPlansTableName + `
		SET
			merchant_id = :merchant_id, 
			acquirer = :acquirer, 
			settlement_type = :settlement_type, 
			installment_type = :installment_type, 
			payment_method = :payment_method, 
			title = :title, 
			description = :description, 
			tenor = :tenor, 
			metadata = :metadata, 
			updated_at = :updated_at
		WHERE 
			uuid = :uuid AND deleted_at IS NULL
	`
	_, err := r.db.NamedExecContext(ctx, query, plan)
	if err != nil {
		r.logger.Error(ctx, "error when inserting installment plan", logger.Error(err), logger.Any("payload", plan))
		return err
	}
	return nil
}
