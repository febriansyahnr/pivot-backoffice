package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"

	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *DisbursementRepository) GetAvgDurationOfBankTransferProcessInMs(ctx context.Context, startTime, endTime time.Time) (ms float64, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetAvgDurationOfBankTransferProcessInMs")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConstant.CtxSQLTableNameKey, "disbursements,outbound")

	rawQuery := `
		SELECT
			IFNULL(AVG(CASE 
				WHEN o.response_time LIKE '%µs' THEN CAST(REPLACE(o.response_time, 'µs', '') AS DECIMAL(20, 6)) / 1000
				WHEN o.response_time LIKE '%ms' THEN CAST(REPLACE(o.response_time, 'ms', '') AS DECIMAL(20, 6))
				WHEN o.response_time LIKE '%s' THEN CAST(REPLACE(o.response_time, 's', '') AS DECIMAL(20, 6)) * 1000
				ELSE 0
			END), 0) AS avg_duration_in_ms
		FROM disbursements d
		JOIN outbound o ON o.origin_id = d.uuid
		WHERE 
			(d.created_at BETWEEN ? AND ?) 
			AND d.status = 'APPROVED'
			AND (o.status_code BETWEEN 200 AND 299)`

	err = r.db.GetContext(ctx, &ms, rawQuery, startTime, endTime)
	return
}

func (r *DisbursementRepository) GetSummaryOfDelayedTransactionBeforeProcessed(ctx context.Context, startTime, endTime time.Time) (result disbursementModel.AfterPayoutCutOffTimeSummary, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetSummaryOfDelayedTransactionBeforeProcessed")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConstant.CtxSQLTableNameKey, "disbursements,account_transactions")

	rawQuery := `
		SELECT
			d.beneficiary_bank_name AS name, COUNT(d.uuid) AS total, SUM(d.amount) AS amount
		FROM disbursements d
		JOIN account_transactions at ON at.reference_id = d.uuid AND at.type = 'DISBURSEMENT'
		WHERE 
			(d.created_at BETWEEN ? AND ?)
			AND d.status = 'APPROVED' AND at.status = 'PENDING' AND at.reason_type = 'CUT_OFF_TIME'
		GROUP BY d.beneficiary_bank_code, d.beneficiary_bank_name`
	if err = r.db.SelectContext(ctx, &result.Banks, rawQuery, startTime, endTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}
		return result, err
	}
	for i := range result.Banks {
		result.Total += result.Banks[i].Total
		result.Amount += result.Banks[i].Amount
	}
	return
}
