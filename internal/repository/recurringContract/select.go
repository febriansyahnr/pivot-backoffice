package recurringContractRepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *repository) GetDetailByID(ctx context.Context, merchantID, uuid string) (*model.RecurringContractDetail, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/recurringContract/GetDetailByID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT
		rc.uuid, rc.merchant_id, rc.customer_id, rc.payment_method_id, rc.payment_token_id, rc.client_reference_id,
		rc.auth_method, rc.auth_transaction_id, rc.plan, rc.trials, rc.billing, rc.currency, rc.amount, rc.status, rc.updated_at, rc.created_at,
		pm.category, pm.type, pm.sub_type, at.processor_reference, at.processor_reference_id, at.processor_transaction_id,
		at.reference_id AS processor_order_id, rc.start_date, rc.end_date
	FROM recurring_contracts rc
	LEFT JOIN payment_methods pm ON pm.uuid = rc.payment_method_id
	LEFT JOIN account_transactions at ON at.uuid = rc.auth_transaction_id AND (at.created_at BETWEEN rc.created_at AND rc.updated_at)
	WHERE
		rc.merchant_id = ? AND rc.uuid = ?;`

	detail := model.RecurringContractDetail{}
	if err := r.db.GetContext(ctx, &detail, rawQuery, merchantID, uuid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	_ = json.Unmarshal(detail.RawPlan, &detail.Plan)
	if detail.RawTrials.Valid {
		_ = json.Unmarshal(detail.RawTrials.JSONText, &detail.Trials)
	}
	_ = json.Unmarshal(detail.RawBilling, &detail.Billing)

	return &detail, nil
}
