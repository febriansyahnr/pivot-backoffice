package disbursementRepository

import (
	"context"
	"fmt"
	"slices"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) GetXbPayoutDashboardInsights(ctx context.Context, request model.GetXbPayoutDashboardInsightRequest) (*model.XbPayoutDashboardInsights, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetXbPayoutDashboardInsights")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	args := []any{
		request.MerchantId, request.StartDate, request.EndDate,
		request.MerchantId, request.StartDate, request.EndDate,
	}
	const rawQuery = `WITH aggregate_transaction_by_status AS (
		SELECT 
			SUM(IF(p.status = 'WAITING' AND p.reason_type = 'WAITING_FOR_CONFIRMATION', 1, 0)) AS waiting_for_confirm_count,
			SUM(IF(p.status = 'APPROVED' AND p.reason_type = 'DOCUMENT_REQUESTED', 1, 0)) AS information_requested_count,
			SUM(IF(p.status = 'APPROVED' AND p.reason_type = 'SUCCESS', 1, 0)) AS success_count,
			SUM(IF(p.status = 'APPROVED' AND p.reason_type = 'SUCCESS', CAST(IFNULL(metadata->>'$.xbDetail.totalAmount', 0) AS DECIMAL(18, 2)), 0)) AS success_total,
			SUM(IF(p.status = 'APPROVED' AND p.reason_type = 'PENDING', 1, 0)) AS pending_count
		FROM disbursements p
		WHERE p.merchant_id = ?
			AND (p.created_at >= ? AND p.created_at <= ?) AND p.currency != 'IDR'
	),
	transaction_by_country AS (
		SELECT
			p.metadata->>'$.xbDetail.beneficiaryData.countryName' AS country,
			SUM(CAST(IFNULL(metadata->>'$.xbDetail.totalAmount', 0) AS DECIMAL(18, 2))) AS volume
		FROM disbursements p
		WHERE p.merchant_id = ?
			AND (p.created_at >= ? AND p.created_at <= ?)
			AND p.currency != 'IDR' AND p.status = 'APPROVED' AND p.reason_type = 'SUCCESS'
		GROUP BY country 
			ORDER BY volume DESC
	),
	ranked_transaction_by_country AS (
		SELECT ROW_NUMBER() OVER (ORDER BY volume DESC) AS rn, country, volume FROM transaction_by_country
	),
	top_transaction_by_volume AS (
		SELECT country, volume FROM ranked_transaction_by_country WHERE rn <= 3
		UNION ALL
		SELECT 'Others', IFNULL(SUM(volume), 0) FROM ranked_transaction_by_country WHERE rn > 3
	)
	SELECT 
		IFNULL(agg.waiting_for_confirm_count, 0) AS waiting_for_confirm_count,
		IFNULL(agg.information_requested_count, 0) AS information_requested_count,
		IFNULL(agg.pending_count, 0) AS pending_count,
		IFNULL(agg.success_count, 0) AS success_count,
		IFNULL(agg.success_total, 0) AS success_total,
		IF(IFNULL(agg.success_count, 0) > 0, 
			JSON_ARRAYAGG(
				JSON_OBJECT(
					"country", top.country,
					"volume", top.volume,
					"percentage", ROUND((top.volume/agg.success_total)*100,2)
				)
			)
		, NULL) AS top_countries_by_volume
	FROM aggregate_transaction_by_status agg
	LEFT JOIN top_transaction_by_volume top ON 1 = 1
	GROUP BY agg.waiting_for_confirm_count, agg.information_requested_count, agg.success_count, agg.success_total;`

	result := &model.XbPayoutDashboardInsights{}
	if err := r.db.GetContext(ctx, result, rawQuery, args...); err != nil {
		r.pdkLogger.Error(ctx, "failed while executing aggregate query", logger.Error(err))
		return nil, err
	}
	if result.RawTopCountriesByVolume.Valid {
		if err := result.RawTopCountriesByVolume.Unmarshal(&result.TopCountriesByVolume); err != nil {
			r.pdkLogger.Error(ctx, "failed while unmarshal json top countries by volume", logger.Error(err))
			return nil, fmt.Errorf("unmarshal json: %v", err)
		}
		result.TopCountriesByVolume = slices.DeleteFunc(result.TopCountriesByVolume, func(p model.XbPayoutTransactionVolumeByCountry) bool {
			return p.Volume == "0.00"
		})
	}
	return result, nil
}
