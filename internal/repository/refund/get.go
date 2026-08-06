package refundRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (r *RefundRepository) ExistsByClientReferenceAndMerchantID(ctx context.Context, clientReferenceID, merchantID string) (bool, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/refund/ExistsByClientReferenceAndMerchantID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, refundTable)

	// Only a SUCCESS refund reserves a client_reference_id. FAILED (like PENDING /
	// WAITING_BANK_TRANSFER) is excluded so a failed refund can be retried with the
	// same reference — consistent with the payment-level dedup that also skips FAILED.
	const query = `SELECT 1 FROM refunds WHERE client_reference_id = ? AND merchant_id = ? AND status NOT IN ('PENDING', 'WAITING_BANK_TRANSFER', 'FAILED') LIMIT 1`

	var exists int
	err := r.db.GetContext(ctx, &exists, query, clientReferenceID, merchantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("failed to check if refund exists by client_reference_id: %s", clientReferenceID))
		return false, err
	}
	return true, nil
}

func (r *RefundRepository) FindByID(ctx context.Context, id string) (*refundModel.Refund, error) {
	query := `
		SELECT uuid, merchant_id, client_reference_id, payment_id, payment_charge_id,
			   currency, amount, status, reason, description, destination_type, method,
			   created_at, updated_at, metadata
		FROM refunds
		WHERE uuid = ?
		LIMIT 1
	`

	var refund refundModel.Refund
	err := r.db.GetContext(ctx, &refund, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("failed to find refund by id: %s", id))
		return nil, err
	}

	if refund.Metadata.Valid {
		_ = json.Unmarshal(refund.Metadata.JSONText, &refund.MetadataObj)
	}

	return &refund, nil
}

// GetRefundByID retrieves a single refund with full details including channel destination
func (r *RefundRepository) GetRefundByID(ctx context.Context, refundID, merchantID string) (*refundModel.RefundResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/refund/GetRefundByID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, refundTable)

	query := `
		SELECT
			r.uuid as id,
			r.merchant_id,
			r.client_reference_id,
			r.payment_id as payment_session_id,
			r.payment_charge_id as charge_id,
			COALESCE(r.amount, 0) as 'amount.value',
			COALESCE(r.currency, "") as 'amount.currency',
			ROUND(COALESCE(charge.credit, 0), 2) as 'captured_amount.value',
			COALESCE(charge.currency, "") as 'captured_amount.currency',
			CASE WHEN COALESCE(r.amount, 0) = COALESCE(charge.credit, 0) THEN true ELSE false END as 'is_full_amount',
			r.status,
			r.reason,
			r.description,
			r.destination_type,
			r.method,
			r.metadata->"$.transferDestination" as 'transfer_destination',
			r.metadata->'$.clientMetadata' as 'metadata',
			COALESCE(rl.reason_type, '') as failure_code,
			charge.channel as payment_channel,
			charge.additional_info as charge_additional_info,
			r.created_at,
			r.updated_at
		FROM
			refunds r
		JOIN account_transactions charge
		ON charge.uuid = r.payment_charge_id AND charge.` + "`type`" + ` = 'PAYMENT'
		JOIN account_transactions rl
		ON rl.reference_id = r.uuid AND rl.` + "`type`" + ` = 'REFUND'
		WHERE r.uuid = ? AND r.merchant_id = ?
		LIMIT 1`

	var result refundModel.RefundResponse
	err := r.db.GetContext(ctx, &result, query, refundID, merchantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when get refund by id", logger.Error(err), logger.String("refundID", refundID))
		return nil, err
	}

	// Process metadata
	result.Metadata, _ = util.ConvertToStruct[map[string]interface{}](result.Metadata)

	// Build channel destination for CHANNEL destination type
	result.BuildChannelDestination()

	return &result, nil
}

// ListByPaymentID retrieves all refund records associated with a specific payment.
// It joins with payments, account_transactions (for charge details), and account_transactions
// (for refund reason) to provide complete refund information. Supports optional status filtering.
// Results are ordered by refund creation date ascending.
func (r *RefundRepository) ListByPaymentID(ctx context.Context, paymentID string, request refundModel.ListByPaymentIDRequest) (result []refundModel.RefundResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payments/GetRefundList")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, refundTable)

	rawQuery := `SELECT
		r.uuid as id, r.merchant_id, r.client_reference_id, r.payment_id as payment_session_id, 
		r.payment_charge_id as charge_id, COALESCE(r.amount, 0) as 'amount.value', COALESCE(r.currency, "") as 'amount.currency',
		ROUND(COALESCE(charge.credit, 0), 2) as 'captured_amount.value', COALESCE(charge.currency, "") as 'captured_amount.currency',
		(COALESCE(r.amount, 0) = COALESCE(charge.credit, 0)) AS 'is_full_amount',
		r.status, r.reason, r.description, r.destination_type, r.method, r.metadata->"$.transferDestination" as 'transfer_destination',
		r.metadata->'$.clientMetadata' as 'metadata', COALESCE(rl.reason_type, '') as failure_code, r.created_at, r.updated_at
	FROM
		payments p
	JOIN refunds r ON r.payment_id = p.uuid
	JOIN account_transactions charge ON charge.uuid = r.payment_charge_id AND charge.created_at >= p.created_at
	JOIN account_transactions rl ON rl.reference_id = r.uuid AND rl.type = 'REFUND' AND rl.created_at >= p.created_at
	WHERE
		p.uuid = ?`

	args := []any{paymentID}
	if request.Status != "" {
		rawQuery += " AND status = ?"
		args = append(args, request.Status)
	}
	rawQuery += " ORDER BY r.created_at"

	if err = r.db.SelectContext(ctx, &result, rawQuery, args...); err != nil {
		return nil, err
	}
	for i, row := range result {
		result[i].Metadata, _ = util.ConvertToStruct[map[string]any](row.Metadata)
	}
	return result, nil
}

func (r *RefundRepository) GetRefundList(ctx context.Context, request refundModel.FilterRefundRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payments/GetRefundList")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, refundTable)

	var (
		result               = make([]*refundModel.RefundResponse, 0)
		totalRecord    int64 = 0
		errG                 = new(errgroup.Group)
		whereStatement string
	)

	if request.Page < 1 {
		request.Page = 1
	}

	if request.PerPage < 1 {
		request.PerPage = 10
	}

	if request.Sort == "" {
		request.Sort = "desc"
	}

	if request.SortBy == "" {
		request.SortBy = "created_at"
	}

	offset := (request.Page - 1) * request.PerPage

	queryTemplate := `
		SELECT
			%s
		FROM
			refunds r
		JOIN account_transactions charge 
		ON charge.uuid = r.payment_charge_id AND charge.` + "`type`" + ` = 'PAYMENT'
		JOIN account_transactions rl
		ON rl.reference_id = r.uuid AND rl.` + "`type`" + ` = 'REFUND'`

	selectStatement := `
		r.uuid as id, 
		r.merchant_id,
		r.client_reference_id, 
		r.payment_id as payment_session_id, 
		r.payment_charge_id as charge_id,
		COALESCE(r.amount, 0) as 'amount.value',
		COALESCE(r.currency, "") as 'amount.currency',
		ROUND(COALESCE(charge.credit, 0), 2) as 'captured_amount.value',
		COALESCE(charge.currency, "") as 'captured_amount.currency',
		CASE WHEN COALESCE(r.amount, 0) = COALESCE(charge.credit, 0) THEN true ELSE false END as 'is_full_amount',
		r.status, 
		r.reason,
		r.description,
		r.destination_type,
		r.method,
		r.metadata->"$.transferDestination" as 'transfer_destination',
		r.metadata->'$.clientMetadata' as 'metadata',
		COALESCE(rl.reason_type, '') as failure_code,
		r.created_at,
		r.updated_at`

	whereClause, args := r.BuildWhereClause(request)
	if len(whereClause) > 0 {
		whereStatement = " WHERE " + strings.Join(whereClause, " AND ")
	}

	errG.Go(func() error {
		query := fmt.Sprintf(queryTemplate, "COUNT(r.uuid)")
		query = query + whereStatement
		err := r.db.GetContext(ctx, &totalRecord, query, args...)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	})

	errG.Go(func() error {
		sortStatement := " ORDER BY r.created_at DESC"
		limitStatement := " LIMIT ? OFFSET ?"
		args2 := append(args, request.PerPage, offset)

		query := fmt.Sprintf(queryTemplate, selectStatement)
		query = query + whereStatement + sortStatement + limitStatement
		err := r.db.SelectContext(ctx, &result, query, args2...)
		return err
	})

	err := errG.Wait()
	if err != nil {
		r.logger.Error(ctx, "error when get refund list", logger.Error(err))
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(totalRecord) / float64(request.PerPage)))

	for _, row := range result {
		row.Metadata, _ = util.ConvertToStruct[map[string]interface{}](row.Metadata)
	}

	return &commonModel.PaginationResponse{
		Data: result,
		Meta: commonModel.Meta{
			Page:       int64(request.Page),
			PerPage:    int64(request.PerPage),
			TotalItems: totalRecord,
			TotalPages: totalPages,
		},
	}, nil
}

func (r *RefundRepository) BuildWhereClause(request refundModel.FilterRefundRequest) ([]string, []any) {
	var whereClause []string
	var args []any

	if request.MerchantID != "" {
		whereClause = append(whereClause, "r.merchant_id = ?")
		args = append(args, request.MerchantID)
	}

	if request.UUID != "" {
		whereClause = append(whereClause, "r.uuid = ?")
		args = append(args, request.UUID)
	}

	if request.PaymentSessionID != "" {
		whereClause = append(whereClause, "r.payment_id = ?")
		args = append(args, request.PaymentSessionID)
	}

	if request.ClientReferenceID != "" {
		whereClause = append(whereClause, "r.client_reference_id = ?")
		args = append(args, request.ClientReferenceID)
	}

	if request.Status != "" {
		whereClause = append(whereClause, "r.status = ?")
		args = append(args, request.Status)
	}

	if request.StartCreatedAt != nil {
		whereClause = append(whereClause, "r.created_at >= ?")
		args = append(args, *request.StartCreatedAt)
	}

	if request.EndCreatedAt != nil {
		whereClause = append(whereClause, "r.created_at <= ?")
		args = append(args, *request.EndCreatedAt)
	}

	return whereClause, args
}

func (r *RefundRepository) FindRefundByChargeID(ctx context.Context, chargeID string) (*refundModel.Refund, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/refund/FindRefundByChargeID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, refundTable)

	const query = `SELECT 
    	uuid, merchant_id, client_reference_id, payment_id, payment_charge_id,
		currency, amount, status, reason, description,
		destination_type, method, created_at, updated_at, metadata
    FROM refunds WHERE payment_charge_id = ? ORDER BY created_at DESC LIMIT 1`

	var refund refundModel.Refund
	err := r.db.GetContext(ctx, &refund, query, chargeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("failed to find refund by payment_charge_id: %s", chargeID))
		return nil, err
	}
	return &refund, nil
}

// GetTotalRefundedAmount retrieves the sum of all refund amounts for a payment,
// excluding refunds with FAILED status. This is more efficient than fetching
// all refund records and summing in application code.
func (r *RefundRepository) GetTotalRefundedAmount(ctx context.Context, paymentID string) (totalAmount float64, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/refund/GetTotalRefundedAmount")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, refundTable)

	const query = `SELECT COALESCE(SUM(amount), 0) FROM refunds WHERE payment_id = ? AND status != 'FAILED'`

	if err = r.db.GetContext(ctx, &totalAmount, query, paymentID); err != nil {
		r.logger.Error(ctx, "failed to get total refunded amount for payment_id: "+paymentID, logger.Error(err))
		return 0, err
	}
	return totalAmount, nil
}
