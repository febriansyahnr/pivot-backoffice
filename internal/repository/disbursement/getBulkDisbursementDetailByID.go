package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *DisbursementRepository) GetBulkDisbursementDetailByID(ctx context.Context, id string) (*disbursementModel.BulkDisbursementDetail, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetBulkDisbursementDetailByID")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "bulk_disbursements")

	var data disbursementModel.BulkDisbursementDetail

	query := `SELECT
		bd.uuid,
		bd.merchant_id,
		bd.status,
		COALESCE(u.name, 'System') AS created_by,
		bd.created_at,
		bd.updated_at,
		bd.deleted_at,
		COALESCE(SUM(d.amount), 0) AS total_amount,
		COUNT(d.uuid) AS total_trx,
		COALESCE(COUNT(d.status = 'APPROVED'), 0) AS total_approved,
		COALESCE(COUNT(d.status = 'REJECTED'), 0) AS total_rejected,
		SUM(
			CASE WHEN 
				d.status = 'APPROVED' AND 
				t.status = 'SUCCESS'
			THEN d.amount 
			ELSE 0 END
		) AS total_success,
		SUM(
			CASE WHEN 
				d.status = 'APPROVED'
				AND t.status = 'FAILED'
			THEN d.amount 
			ELSE 0 END
		) AS total_failed,
		SUM(
			CASE WHEN d.reason_type = 'CANCELLED' THEN d.amount ELSE 0 END
		) AS total_cancelled,
		SUM(
			CASE WHEN d.status = 'APPROVED' AND t.status = 'PENDING' THEN d.amount ELSE 0 END
		) AS total_pending
	FROM bulk_disbursements bd
	LEFT JOIN users u
		ON u.uuid = bd.created_by
	LEFT JOIN disbursements d
		ON d.bulk_id = bd.uuid
	LEFT JOIN account_transactions t
		ON t.reference_id = d.uuid
		AND t.type = 'DISBURSEMENT'
		AND IFNULL(t.reason_type, '') <> 'REVERSAL'
	WHERE bd.uuid = ?
	GROUP BY
		bd.uuid,
		bd.merchant_id,
		bd.status,
		u.name,
		bd.created_at,
		bd.updated_at,
		bd.deleted_at;`

	if err := r.db.GetContext(ctx, &data, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &data, nil
}
