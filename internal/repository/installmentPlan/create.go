package installmentPlan

import (
	"context"

	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *InstallmentPlanRepository) Create(ctx context.Context, plan *installmentPlanModel.InstallmentPlan) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/installmentPlan/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, installmentPlansTableName)
	query := `
		INSERT INTO ` + installmentPlansTableName + `
			(uuid, merchant_id, acquirer, settlement_type, installment_type, payment_method, title, description, tenor, status, metadata, created_at, updated_at) 
		VALUES (:uuid, :merchant_id, :acquirer, :settlement_type, :installment_type, :payment_method, :title, :description, :tenor, :status, :metadata, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, plan)
	if err != nil {
		r.logger.Error(ctx, "error when inserting installment plan", logger.Error(err), logger.Any("payload", plan))
		return err
	}
	return nil
}
