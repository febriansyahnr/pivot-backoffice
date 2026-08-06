package disbursementRepository

import (
	"context"
	"fmt"
	"time"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *DisbursementRepository) SumAmountByIDs(ctx context.Context, ids []string) (*disbursementModel.SumAmountResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/SumAmountByIDs")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "disbursements")

	var sumAmountResponse disbursementModel.SumAmountResponse

	getFinalFee := `IF(IFNULL(metadata->>'$.feeDetail.deductionType', 'DIRECT') = 'DIRECT', IFNULL(metadata->>'$.feeDetail.finalAmount', 0), 0)`
	queryCount := fmt.Sprintf(`SELECT 
		SUM(IF(metadata->>'$.onBehalf.parentMerchantId' IS NULL, (amount + %s), total_amount)) AS sum_total_amount, SUM(%s) AS sum_parent_fee_charged
	FROM disbursements`, getFinalFee, getFinalFee)

	if len(ids) > 0 {
		idString := fmt.Sprintf("'%s'", ids[0])
		for _, id := range ids[1:] {
			idString += fmt.Sprintf(", '%s'", id)
		}

		queryCount += fmt.Sprintf(" WHERE uuid IN (%s)", idString)

		err := r.db.GetContext(ctx, &sumAmountResponse, queryCount)
		if err != nil {
			r.pdkLogger.Error(ctx, "error when sum disbursements", logger.Error(err))
			return nil, err
		}
	}

	return &sumAmountResponse, nil
}

func (r *DisbursementRepository) GetBeneficiaryTransactionLimit(ctx context.Context, merchantId, bankCode, accountNo string, startAt, endAt time.Time) (*disbursementModel.BeneficiaryPayoutLimitRuleLimit, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/disbursement/GetBeneficiaryTransactionLimit")
	defer span.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "disbursements")

	var response disbursementModel.BeneficiaryPayoutLimitRuleLimit

	queryCount := `SELECT 
		COUNT(d.uuid) AS count_payout,
		COALESCE(SUM(d.amount), 0) AS processed
	FROM disbursements d 
	JOIN merchants m
		ON d.merchant_id = m.uuid
	LEFT JOIN merchants m2
		On d.merchant_id = m2.uuid AND m2.kyc_status = "NOT_REQUIRED"
	JOIN account_transactions t 
		ON t.reference_id = d.uuid 
		AND t.type = 'DISBURSEMENT'
		AND t.status IN ('PENDING', 'SUCCESS')
		AND (t.updated_at BETWEEN ? AND ?)
		AND (t.created_at BETWEEN DATE_SUB(?, INTERVAL 30 DAY) AND ?)
		AND t.deleted_at IS NULL
	WHERE
		d.beneficiary_bank_code = ? AND d.beneficiary_account_no = ?
		AND d.updated_at > ? AND d.type = 'LOCAL_PAYOUT'`

	// 72 hours updated disbursement
	disbursementUpdatedAt := time.Now().UTC().Add(-72 * time.Hour)

	var err error
	if merchantId != "" {
		queryCount += ` AND (d.merchant_id = ? OR m2.parent_id = ?)`
		err = r.db.GetContext(ctx, &response, queryCount, startAt, endAt, startAt, endAt, bankCode, accountNo, disbursementUpdatedAt, merchantId, merchantId)
	} else {
		err = r.db.GetContext(ctx, &response, queryCount, startAt, endAt, startAt, endAt, bankCode, accountNo, disbursementUpdatedAt)
	}

	if err != nil {
		r.pdkLogger.Error(ctx, "error when get beneficiary transaction limit", logger.Error(err))
		return nil, err
	}

	return &response, nil
}
