package accounttransaction_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountTransactionRepository) UpdateBulkReconStatus(ctx context.Context, params *reconciliation.BulkUpatedStatus) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateBulkReconStatus")
	defer segment.End()
	// get time duration from statTime and endTime
	startTime := params.StartTime
	endTime := params.EndTime
	duration := endTime.Sub(startTime)

	// if duration > 3 days, set start time to endtime - 3 days
	// max recon time is 3 days
	if duration > 3*24*time.Hour {
		startTime = endTime.Add(-3 * 24 * time.Hour)
	}

	// get date from params wth format YYYY-MM-DD
	startDateStr := startTime.Format("2006-01-02")
	endDateStr := endTime.Format("2006-01-02")

	startUpdatedAt := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location()) // Set update_at to beginning of day
	endUpdatedAt := startUpdatedAt.Add(time.Duration(params.ScanningToleranceInDays) * 24 * time.Hour)

	query := fmt.Sprintf(`
		UPDATE account_transactions 
		SET additional_info = JSON_SET(additional_info, '$.reconDetail', JSON_OBJECT("status", "%s"))
		WHERE 
		updated_at >= ? AND updated_at < ?
		AND date(transaction_timestamp) BETWEEN '%s' and '%s'
		AND status IN ('SUCCESS', 'PENDING')
		AND type = ?
		AND reference = ? 
		AND (JSON_EXTRACT(additional_info, "$.reconDetail") is null 
		     OR JSON_EXTRACT(additional_info, "$.reconDetail.status") != 'TRUE');
	`, params.Status, startDateStr, endDateStr)

	_, err := r.db.ExecContext(
		ctx,
		query,
		startUpdatedAt,
		endUpdatedAt,
		params.TrxType,
		params.TrxReference,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating status account_transactions for recon",
			logger.Error(err),
			logger.Any("request", params),
			logger.Any("query", query))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) SetAdditionalInfoReconciliation(ctx context.Context, id string, params *reconciliation.ReconDetail) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/SetAdditionalInfoReconciliation")
	defer segment.End()

	var addInfo json.RawMessage
	addInfo, _ = json.Marshal(params)

	query := fmt.Sprintf(`
		UPDATE account_transactions
		SET additional_info =  JSON_MERGE_PATCH(COALESCE(additional_info, '{}'),'{"reconDetail": %s}')
		WHERE uuid = ?;
	`, addInfo)

	_, err := r.db.ExecContext(
		ctx,
		query,
		id,
	)
	if err != nil && !errors.Is(err, constant.ErrNoRowsAffected) {
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) UpdateReconDetail(ctx context.Context, id string, params *reconciliation.ReconDetail) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateReconDetail")
	defer segment.End()

	addQuery := ""
	if params.DateTime != "" {
		addQuery = fmt.Sprintf(", '$.reconDetail.datetime', '%s'", params.DateTime)
	}

	query := fmt.Sprintf(`
		UPDATE account_transactions
		SET
			additional_info = JSON_SET(
				additional_info 
				,'$.reconDetail.status', ? 
				,'$.reconDetail.reason', ?
				%s
			)
		WHERE uuid = ?
	`, addQuery)

	_, err := r.db.ExecContext(
		ctx,
		query,
		params.Status,
		params.Reason,
		id,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating recon detail ",
			logger.Error(err),
			logger.Any("request", params),
			logger.Any("query", query))
		return err
	}

	return nil
}
