package merchant

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

const onBehalfFeeConfigTableName = "on_behalf_fee_configs"

func (r *MerchantRepository) ValidateCreateFeeConfigOnBehalf(ctx context.Context, request *merchant.CreateFeeConfigOnBehalfRequest) (bool, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/ValidateCreateFeeConfigOnBehalf")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchants,on_behalf_fee_configs")

	args, point := []interface{}{request.MerchantId}, 1
	query := `SELECT COUNT(uuid) AS point FROM merchants WHERE uuid = ?`

	if request.Type == constant.FeeOnBehalfTypeDirect {
		point += 2

		args = append(args, request.SubMerchantId, request.MerchantId)
		query += ` UNION ALL SELECT COUNT(uuid) FROM merchants WHERE uuid = ? AND parent_id = ?`

		args = append(args, request.MerchantId, request.SubMerchantId, request.Reference)
		query += ` UNION ALL SELECT COUNT(id) = 0 FROM on_behalf_fee_configs WHERE merchant_id = ? AND sub_merchant_id = ? AND reference = ? AND deleted_at IS NULL`

		if request.Reference == constant.ReferencePayment {
			query += ` AND payment_method = ?`
			args = append(args, request.PaymentMethod)
		}

	} else {

		point += 1
		args = append(args, request.MerchantId, request.Reference)
		query += ` UNION ALL SELECT COUNT(id) = 0 FROM on_behalf_fee_configs WHERE merchant_id = ? AND reference = ? AND deleted_at IS NULL`

		if request.Reference == constant.ReferencePayment {
			query += ` AND payment_method = ?`
			args = append(args, request.PaymentMethod)
		}
		if request.Type == constant.FeeOnBehalfTypeDefault {
			query += ` AND type = 'DEFAULT'`
		}
		if request.ReferenceType != "" {
			query += ` AND reference_type = ?`
			args = append(args, request.ReferenceType)
		}
	}

	point += 1
	args = append(args, request.MerchantId, request.Reference)
	query += ` UNION ALL SELECT COUNT(id) = 0 FROM on_behalf_fee_configs WHERE merchant_id = ? AND reference = ? AND deleted_at IS NULL`
	if request.Type == constant.FeeOnBehalfTypeAll {
		query += ` AND type IN ('DEFAULT', 'DIRECT')`

	} else {
		query += ` AND type = 'ALL'`
	}
	if request.Reference == constant.ReferencePayment {
		query += ` AND payment_method = ?`
		args = append(args, request.PaymentMethod)
	}
	if request.ReferenceType != "" {
		query += ` AND reference_type = ?`
		args = append(args, request.ReferenceType)
	}

	result := 0
	err := r.db.GetContext(ctx, &result, "SELECT SUM(point) FROM ("+query+") foo", args...)

	return result == point, err
}

func (r *MerchantRepository) CreateFeeConfigOnBehalf(ctx context.Context, data *merchant.OnBehalfFeeConfig) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/CreateFeeConfigOnBehalf")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, onBehalfFeeConfigTableName)

	rawQuery := `INSERT INTO on_behalf_fee_configs
			(id, merchant_id, type, sub_merchant_id, reference, reference_type, payment_method, amount_type, amount, percentage, created_at, updated_at)
		VALUES(:id, :merchant_id, :type, :sub_merchant_id, :reference, :reference_type, :payment_method, :amount_type, :amount, :percentage, :created_at, :updated_at);`

	_, err := r.db.NamedExecContext(ctx, rawQuery, data)
	return err
}

func (r *MerchantRepository) GetFeeConfigOnBehalf(ctx context.Context, request *merchant.GetFeeConfigOnBehalfRequest) ([]merchant.FeeConfigOnBehalfResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetFeeConfigOnBehalf")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, onBehalfFeeConfigTableName)

	rawQuery := `SELECT
			id, type, sub_merchant_id, amount_type, amount, percentage, created_at, updated_at
		FROM on_behalf_fee_configs WHERE merchant_id = ? AND reference = ?`
	args := []interface{}{request.MerchantId, request.Reference}

	if request.Reference == constant.ReferencePayment {
		rawQuery += ` AND payment_method = ?`
		args = append(args, request.PaymentMethod)
	}
	if request.ReferenceType != "" {
		rawQuery += ` AND reference_type = ?`
		args = append(args, request.ReferenceType)
	}
	rawQuery += ` AND deleted_at IS NULL ORDER BY CASE WHEN type = 'DIRECT' THEN 1 ELSE 2 END`

	result := []merchant.FeeConfigOnBehalfResponse{}
	if err := r.db.SelectContext(ctx, &result, rawQuery, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return result, nil
}

func (r *MerchantRepository) UpdateFeeConfigOnBehalf(ctx context.Context, id string, request *merchant.UpdateFeeConfigOnBehalfRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/UpdateFeeConfigOnBehalf")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, onBehalfFeeConfigTableName)

	rawQuery := `UPDATE on_behalf_fee_configs
		SET
			amount_type = ?, amount = ?, percentage = ?, updated_at = ?
		WHERE id = ?;`
	args := []interface{}{
		request.AmountType, request.Amount, request.Percentage, time.Now().UTC(), id,
	}
	if affected, err := r.db.ExecContext(ctx, rawQuery, args...); err != nil {
		return err

	} else if !affected {
		return constant.ErrNoRowsAffected
	}
	return nil
}

func (r *MerchantRepository) GetTransactionFeeOnBehalf(ctx context.Context, merchantId, subMerchantId, reference, paymentMethod, referenceType string) (*merchant.TransactionFeeOnBehalf, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetTransactionFeeOnBehalf")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, onBehalfFeeConfigTableName)

	masterQuery := `SELECT
			reference, type, amount_type, amount, percentage 
		FROM on_behalf_fee_configs
		WHERE
			merchant_id = ? AND deleted_at IS NULL`

	query := masterQuery + ` AND type IN ('ALL', 'DEFAULT') AND reference = ?`
	args := []interface{}{merchantId, reference}

	if reference == constant.ReferencePayment {
		query += ` AND payment_method = ?`
		args = append(args, paymentMethod)
	}
	if referenceType != "" {
		query += ` AND reference_type = ?`
		args = append(args, referenceType)
	}

	query += ` UNION ALL ` + masterQuery + ` AND type = 'DIRECT' AND reference = ? AND sub_merchant_id = ?`
	args = append(args, merchantId, reference, subMerchantId)

	if reference == constant.ReferencePayment {
		query += ` AND payment_method = ?`
		args = append(args, paymentMethod)
	}
	query += ` ORDER BY CASE WHEN type = 'DIRECT' THEN 1 ELSE 2 END LIMIT 1`

	result := &merchant.TransactionFeeOnBehalf{}
	if err := r.db.GetContext(ctx, result, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}
