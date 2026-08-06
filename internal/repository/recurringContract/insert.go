package recurringContractRepo

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *repository) Insert(ctx context.Context, data model.RecurringContract) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/recurringContract/Insert")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	data.RawPlan, _ = json.Marshal(data.Plan)
	data.RawBilling, _ = json.Marshal(data.Billing)
	if len(data.Trials) > 0 {
		data.RawTrials.Valid = true
		data.RawTrials.JSONText, _ = json.Marshal(data.Trials)
	}

	rawQuery := `INSERT INTO recurring_contracts (
		uuid, merchant_id, client_reference_id, customer_id, auth_method, plan, trials, billing, scheduler_mode, currency, amount, status, created_by, created_at, updated_by, updated_at, end_date
	) VALUES(
		:uuid, :merchant_id, :client_reference_id, :customer_id, :auth_method, :plan, :trials, :billing, :scheduler_mode, :currency, :amount, :status, :created_by, :created_at, :updated_by, :updated_at, :end_date
	);`

	if _, err := r.db.NamedExecContext(ctx, rawQuery, data); err != nil {
		if strings.Contains(err.Error(), "Error 1062 (23000): Duplicate entry") {
			return constant.ErrClientReferenceIDAlreadyExist
		}
		return err
	}
	return nil
}
