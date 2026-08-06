package recurringContractRepo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *repository) UpdateRecurringContractStatus(ctx context.Context, uuid, status, updatedBy string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/recurringContract/UpdateRecurringContractStatus")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	previousStatus := ""
	fieldsToBeUpdated := "status = ?, updated_at = ?"
	args := []any{status, time.Now().UTC()}

	if updatedBy != "" {
		fieldsToBeUpdated += ", updated_by = ?"
		args = append(args, updatedBy)
	}

	switch status {
	case constant.RecurringContractStatusPendInitialAuth:
		previousStatus = fmt.Sprintf("status IN ('%s', '%s')", constant.RecurringContractStatusCreated, constant.RecurringContractStatusPendInitialAuth)

	case constant.RecurringContractStatusActive:
		fieldsToBeUpdated += ", activated_at = ?"
		args = append(args, time.Now().UTC())
		previousStatus = fmt.Sprintf("status = '%s'", constant.RecurringContractStatusPendInitialAuth)

	case constant.RecurringContractStatusInactive:
		fieldsToBeUpdated += ", deactivated_at = ?"
		args = append(args, time.Now().UTC())
		previousStatus = fmt.Sprintf("status IN ('%s', '%s', '%s')", constant.RecurringContractStatusCreated, constant.RecurringContractStatusPendInitialAuth, constant.RecurringContractStatusActive)

	default:
		return errors.New("invalid or unregistered status")
	}

	rawQuery := fmt.Sprintf(`UPDATE 
			recurring_contracts
		SET
			%s
		WHERE uuid = ? AND %s;`, fieldsToBeUpdated, previousStatus,
	)
	args = append(args, uuid)

	if affected, err := r.db.ExecContext(ctx, rawQuery, args...); err != nil {
		return err

	} else if !affected {
		return constant.ErrNoRowsAffected
	}
	return nil
}

func (r *repository) UpdateRecurringContract(ctx context.Context, payload model.UpdateRecurringContractRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/recurringContract/UpdateRecurringContract")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	params, values := []string{}, []any{}

	if payload.BillingCycleCount > 0 {
		values = append(values, payload.BillingCycleCount)
		params = append(params, "billing = JSON_SET(billing, '$.count', ?)")
	}
	if payload.Status == constant.RecurringContractStatusActive {
		values = append(values, payload.Status, payload.ActivatedAt)
		params = append(params, "status = ?", "activated_at = ?")
	}
	if payload.TransactionID != "" {
		values = append(values, payload.TransactionID)
		params = append(params, "auth_transaction_id = ?")
	}
	if payload.PaymentTokenID != "" {
		values = append(values, payload.PaymentTokenID)
		params = append(params, "payment_token_id = ?")
	}
	if payload.PaymentMethodID != "" {
		values = append(values, payload.PaymentMethodID)
		params = append(params, "payment_method_id = ?")
	}
	if !payload.UpdatedAt.IsZero() {
		values = append(values, payload.UpdatedAt)
		params = append(params, "updated_at = ?")
	}
	if payload.UpdatedBy != "" {
		values = append(values, payload.UpdatedBy)
		params = append(params, "updated_by = ?")
	}

	if len(params) == 0 {
		return errors.New("update parameters must not be empty")
	}

	args := append(values, payload.RecurringID)
	rawQuery := fmt.Sprintf(`UPDATE recurring_contracts SET %s WHERE uuid = ?;`, strings.Join(params, ", "))

	if affected, err := r.db.ExecContext(ctx, rawQuery, args...); err != nil {
		return err

	} else if !affected {
		return constant.ErrNoRowsAffected
	}
	return nil
}
