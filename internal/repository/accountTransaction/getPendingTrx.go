package accounttransaction_repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx/types"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *AccountTransactionRepository) GetListOfTransactionReferenceIdsWithPendingStatus(ctx context.Context, merchantId, accountId string, startTime, endTime time.Time) (result []string, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/GetListOfTransactionReferenceIdsWithPendingStatus")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT 
		IFNULL(JSON_ARRAYAGG(reference_id), JSON_ARRAY())
	FROM (
		SELECT
			reference_id
		FROM account_transactions
		WHERE
			merchant_id = ? AND account_id = ? 
			AND (
				type != 'FEE'
				OR (
					additional_info->>'$.deductionType' != 'MANUAL' 
					AND (
						additional_info->>'$.deductionType' = 'DIRECT'
						OR ( additional_info->>'$.deductionType' = 'AUTOMATED' AND status = 'SUCCESS' )
					)
				)
			) AND status = 'PENDING' AND settlement_status IS NULL
			AND (updated_at >= ? AND updated_at < ?)
		GROUP BY reference_id
	) foo`

	rawJSON := types.JSONText{}
	if err = r.db.GetContext(ctx, &rawJSON, rawQuery, merchantId, accountId, startTime, endTime); err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawJSON, &result)
	return
}
