package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cardFundedPayout"
	statusHistoriesModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/statusHistory"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

// SelectCardFundedPayoutDetailStr is the select string for card-funded payout detail
const SelectCardFundedPayoutDetailStr = `d.uuid, d.reference_id, d.created_at, d.amount, IFNULL(d.fee, '') AS fee, d.total_amount,
	d.status, IFNULL(d.beneficiary_bank_name, '') AS beneficiary_bank_name, d.beneficiary_account_no, d.beneficiary_account_name,
	IFNULL(d.remark, '') AS remark, d.approved_at, IFNULL(u.name, '') as approved_by, d.metadata,
	IFNULL(t.status, '') AS transaction_status, d.merchant_id`

func (r *DisbursementRepository) GetCardFundedPayoutDetail(
	ctx context.Context,
	filter *cardFundedPayoutModel.GetPayoutDetailRequest,
) (*cardFundedPayoutModel.GetPayoutDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/GetCardFundedPayoutDetail")
	defer segment.End()
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `SELECT ` + SelectCardFundedPayoutDetailStr + `
		FROM disbursements d
		LEFT JOIN account_transactions t ON d.uuid = t.reference_id AND t.type = '` + constant.TypeDisbursement + `' AND IFNULL(t.reason_type, '') != 'REVERSAL'
		LEFT JOIN users u ON d.approved_by = u.uuid
		WHERE d.uuid = ? AND d.type = '` + constant.DisbursementTypeCardFundedPayout + `' AND d.merchant_id = ? AND d.deleted_at IS NULL`

	var data cardFundedPayoutModel.GetPayoutDetailResponse
	if err := r.db.GetContext(ctx, &data, query, filter.PayoutID, filter.MerchantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Fetch status history
	statusHistories, err := r.getStatusHistoryByReference(ctx, data.UUID, constant.DisbursementTypeCardFundedPayout)
	if err != nil {
		return nil, err
	}

	// Convert status histories to response format
	data.StatusHistory = make([]cardFundedPayoutModel.StatusHistoryItem, 0, len(statusHistories))
	for _, history := range statusHistories {
		item := cardFundedPayoutModel.StatusHistoryItem{
			Status:    history.Status,
			Timestamp: &history.CreatedAt,
		}
		if history.MetadataObj != nil {
			item.Label = history.MetadataObj.Label
			item.Description = history.MetadataObj.Description
		}
		data.StatusHistory = append(data.StatusHistory, item)
	}

	data.ChargeIDs, err = r.getChargeIDsByReference(ctx, filter.PayoutID, filter.MerchantID)
	if err != nil {
		return nil, err
	}

	// Hydrate derived fields from metadata
	data.Hydrate()

	return &data, nil
}

func (r *DisbursementRepository) getStatusHistoryByReference(
	ctx context.Context,
	referenceID, referenceType string,
) ([]*statusHistoriesModel.StatusHistory, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/getStatusHistoryByReference")
	defer segment.End()

	query := `SELECT id, reference_id, reference_type, status, metadata, created_at
		FROM status_histories
		WHERE reference_id = ? AND reference_type = ?
		ORDER BY created_at ASC`

	var histories []*statusHistoriesModel.StatusHistory
	if err := r.db.SelectContext(ctx, &histories, query, referenceID, referenceType); err != nil {
		return nil, err
	}

	// Unmarshal metadata for each history
	for _, history := range histories {
		if history.Metadata.Valid {
			_ = history.Metadata.Unmarshal(&history.MetadataObj)
		}
	}

	return histories, nil
}

func (r *DisbursementRepository) getChargeIDsByReference(
	ctx context.Context,
	referenceID, merchantID string,
) ([]string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/disbursement/getChargeIDsByReference")
	defer segment.End()

	query := `SELECT at.uuid 
		FROM payments p 
		LEFT JOIN account_transactions at ON at.type = '` + constant.TypePayment + `' AND p.uuid = at.reference_id
		WHERE p.merchant_id = ? AND p.reference_id = ? 
		ORDER BY p.created_at ASC`

	var chargeIDs []string
	if err := r.db.SelectContext(ctx, &chargeIDs, query, merchantID, referenceID); err != nil {
		return nil, err
	}

	return chargeIDs, nil
}
