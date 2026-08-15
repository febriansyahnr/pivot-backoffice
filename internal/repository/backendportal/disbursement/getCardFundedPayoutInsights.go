package disbursementRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// GetCardFundedPayoutInsights returns the total amount and total transaction count
// for card-funded payouts with WAITING approval status for a given merchant.
func (r *DisbursementRepository) GetCardFundedPayoutInsights(
	ctx context.Context,
	filter *cardFundedPayoutModel.FilterGetPayoutInsights,
) (*cardFundedPayoutModel.GetPayoutInsightsDTO, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetCardFundedPayoutInsights")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `SELECT
		COUNT(uuid) AS count,
		COALESCE(SUM(amount), 0) AS sum
	FROM disbursements
	WHERE merchant_id = ?
		AND type = ?
		AND status = ?
		AND deleted_at IS NULL`

	args := []interface{}{filter.MerchantID, constant.DisbursementTypeCardFundedPayout, constant.DisbursementStatusWaiting}

	if filter.StartCreatedAt != nil && filter.EndCreatedAt != nil {
		query += " AND created_at >= ? AND created_at <= ?"
		args = append(args, filter.StartCreatedAt, filter.EndCreatedAt)
	}

	result := &cardFundedPayoutModel.GetPayoutInsightsDTO{}
	if err := r.db.GetContext(ctx, result, query, args...); err != nil {
		r.pdkLogger.Error(ctx, "failed to get card funded payout insights", logger.Error(err))
		return nil, err
	}

	return result, nil
}
