package merchant

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

func (r *MerchantRepository) mappingFeeUsecase(input string) string {
	if len(input) == 0 {
		return "others"
	}

	switch input {
	case constant.ReferenceDisbursement:
		return "payouts"

	case constant.ReferencePayment:
		return "payments"

	default:
		result := strings.ReplaceAll(util.ToTitle(input), " ", "")
		return strings.ToLower(string(result[0])) + result[1:]
	}
}

func (r *MerchantRepository) GetBillingFees(ctx context.Context, request merchant.BillingFeeRequest) (*merchant.BillingFeeResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetBillingFees")
	defer segment.End()

	response := &merchant.BillingFeeResponse{
		Details: map[string][]merchant.BillingFeeDetailResponse{},
	}
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "account_transactions")

	// will try to get the disbursement and payment
	// as a reference to the sub merchant id
	rawQuery := `
		SELECT 
			at.additional_info->>'$.type' AS type,
			at.additional_info->>'$.method' AS method,
			IFNULL(at.additional_info->>'$.channel', '') AS channel,
			at.additional_info->>'$.amountType' AS fee_type,
			at.additional_info->>'$.amount' AS fee_amount,
			at.additional_info->>'$.percentage' AS fee_percentage,
			COALESCE(p.merchant_id, d.merchant_id, at.merchant_id) AS merchant_id,
			COUNT(at.uuid) AS total, SUM(IFNULL(at.additional_info->>'$.trxAmount', 0)) AS trx_amount, SUM(at.debit) AS total_fee_amount
		FROM
			account_transactions at
		LEFT JOIN payments p ON at.reference_id = p.uuid 
		LEFT JOIN disbursements d ON at.reference_id = d.uuid
		WHERE
			at.merchant_id = ?
			AND (at.updated_at BETWEEN ? AND ?)
			AND at.type = 'FEE' AND at.additional_info->>'$.deductionType' = 'MANUAL' 
			AND IFNULL(at.additional_info->>'$.canceledAt', '') = '' AND at.debit > 0 AND at.status = 'SUCCESS'
			AND IFNULL(at.settlement_status, 'PENDING') = ?
		GROUP BY
			at.additional_info->>'$.type',
			at.additional_info->>'$.method', 
			IFNULL(at.additional_info->>'$.channel', ''),
			at.additional_info->>'$.amountType',
			at.additional_info->>'$.amount',
			at.additional_info->>'$.percentage',
			COALESCE(p.merchant_id, d.merchant_id, at.merchant_id)
		ORDER BY 
			type, method, channel;`

	args := []any{
		request.MerchantId, request.StartDate, request.EndDate, request.Status,
	}
	result := []merchant.BillingFeeDetailResponse{}
	if err := r.db.SelectContext(ctx, &result, rawQuery, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, nil
		}
		r.logger.Error(ctx, "Failed while recap merchant billing fees", logger.Error(err))
		return nil, err
	}

	for _, group := range result {
		key := r.mappingFeeUsecase(group.Type)

		response.Total += group.Total
		response.TotalFeeAmount += group.TotalFeeAmount
		response.Details[key] = append(response.Details[key], group)
	}
	return response, nil
}

func (r *MerchantRepository) PayBillingFees(ctx context.Context, request merchant.PayBillingFeeRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/PayBillingFees")
	defer segment.End()

	rawQuery := `UPDATE
			account_transactions
		SET
			settlement_status = ?, settlement_at = ?, additional_info = JSON_SET(additional_info, '$.paidBy', ?)
		WHERE
			merchant_id = ?
			AND (updated_at BETWEEN ? AND ?)
			AND type = 'FEE' AND additional_info->>'$.deductionType' = 'MANUAL' 
			AND IFNULL(additional_info->>'$.canceledAt', '') = '' AND debit > 0 AND status = 'SUCCESS'
			AND IFNULL(settlement_status, 'PENDING') = 'PENDING';`

	args := []any{
		constant.StatusSuccess, time.Now().UTC(), request.Username, request.MerchantId, request.StartDate, request.EndDate,
	}
	if _, err = r.db.ExecContext(ctx, rawQuery, args...); err != nil {
		r.logger.Error(ctx, "Failed while execute query for pay billing fees", logger.Error(err))
	}
	return err
}
