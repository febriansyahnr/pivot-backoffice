package disbursementRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/jmoiron/sqlx/types"
)

func (r *DisbursementRepository) GetTransactionConfig(ctx context.Context, merchantId string) (config *disbursementModel.TransactionConfig, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetTransactionConfig")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchants")

	var rawConfig types.NullJSONText

	rawQuery := `SELECT 
			IF(
				transaction_configs->>'$.disbursement' IS NULL AND NULLIF(parent_id, '') IS NOT NULL, 
				(SELECT transaction_configs->>'$.disbursement' FROM merchants WHERE uuid = m.parent_id), transaction_configs->>'$.disbursement'
			) AS transaction_configs 
		FROM merchants m WHERE uuid = ? AND deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, &rawConfig, rawQuery, merchantId); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err

	} else if !rawConfig.Valid {
		return &disbursementModel.TransactionConfig{
			MinAmount: constant.DisbursementMinAmount,
			MaxAmount: constant.DisbursementMaxAmount,
		}, nil
	}

	config = &disbursementModel.TransactionConfig{}
	if err = json.Unmarshal(rawConfig.JSONText, config); err != nil {
		return nil, err
	}
	return
}

func (r *DisbursementRepository) GetDailyTransactionLimit(ctx context.Context, merchantId, merchantType string) (*disbursementModel.DailyTransactionLimitResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetDailyTransactionLimit")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "merchants,account_transactions")

	wit := time.Now().In(local)
	startDate := time.Date(wit.Year(), wit.Month(), wit.Day(), 0, 0, 0, 0, local).In(time.UTC)
	endDate := time.Date(wit.Year(), wit.Month(), wit.Day(), 23, 59, 59, 999, local).In(time.UTC)

	result := &disbursementModel.DailyTransactionLimitResponse{}
	rawQuery := `SELECT
		m.transaction_configs->>'$.dailyDisbursement.merchant' AS "limit",
		IFNULL(
			(SELECT
				SUM(debit) 
			FROM account_transactions
			WHERE merchant_id = m.uuid 
				AND (updated_at BETWEEN ? AND ?)
				AND (created_at BETWEEN DATE_SUB(?, INTERVAL 30 DAY) AND ?)
				AND type = 'DISBURSEMENT' AND status IN ('PENDING', 'SUCCESS')
				AND channel != 'XB' AND reference = 'DISBURSEMENT' AND deleted_at IS NULL
			), 0) AS processed
	FROM merchants m
	WHERE 
		m.uuid = ? AND m.deleted_at IS NULL;`

	if merchantType == constant.DisbursementDailyLimitMerchantPlatform {
		rawQuery = `SELECT
			m.transaction_configs->>'$.dailyDisbursement.merchantPlatform' AS "limit",
			IFNULL(
				(SELECT
					SUM(at.debit)
				FROM merchants sub
				JOIN account_transactions at 
					ON at.merchant_id = sub.uuid 
						AND (at.updated_at BETWEEN ? AND ?)
						AND (at.created_at BETWEEN DATE_SUB(?, INTERVAL 30 DAY) AND ?)
						AND at.type = 'DISBURSEMENT' AND at.status IN ('PENDING', 'SUCCESS')
						AND at.channel != 'XB' AND at.reference = 'DISBURSEMENT' AND at.deleted_at IS NULL
				WHERE
					sub.parent_id = m.uuid AND sub.kyc_status = 'NOT_REQUIRED' AND sub.deleted_at IS NULL
				), 0) AS processed
		FROM merchants m
		WHERE 
			m.uuid = ? AND m.deleted_at IS NULL;`
	}

	if err := r.db.GetContext(ctx, result, rawQuery, startDate, endDate, startDate, endDate, merchantId); err != nil {
		return nil, err
	}

	if result.Limit == nil && merchantType == constant.DisbursementDailyLimitMerchant {
		result.Limit = util.ValueToPtr(r.config.DailyLimitMerchant) // Default daily transaction limit for merchant

	} else if result.Limit == nil {
		result.Limit = util.ValueToPtr(r.config.DailyLimitMerchantPlatform) // Default daily transaction limit for merchant platfrom
	}
	result.Remaining = *result.Limit - result.Processed
	return result, nil
}
