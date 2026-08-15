package paymentRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	constant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"

	"github.com/jmoiron/sqlx"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (r *PaymentRepository) GetInvestigatedPayments(
	ctx context.Context,
	filter *paymentModel.GetInvestigatedPaymentsFilterRequest,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/GetInvestigatedPayments")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	var (
		data = make([]*paymentModel.InvestigatedPaymentDTO, 0)
		errG = new(errgroup.Group)
	)

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	offset := (filter.Page - 1) * filter.Limit

	conditions, conditionArgs := buildInvestigationCondition(filter)
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var partitionCondition string
	var args []any
	if filter.FromDate != nil {
		partitionCondition = " AND p.created_at >= DATE_SUB(?, INTERVAL 60 DAY)"
		args = append(conditionArgs, filter.FromDate)
	} else {
		partitionCondition = " AND p.created_at >= DATE_SUB(NOW(), INTERVAL 60 DAY)"
		args = conditionArgs
	}

	query := `
		SELECT
			p.uuid,
			p.reference_id,
			p.amount,
			p.currency,
			p.merchant_id,
			m.name as merchant_name,
			pm.type as payment_method_type,
			pm.name as payment_channel,
			p.status,
			p.reason_type,
			p.reason_description,
			p.updated_at,
			investigation_started_at AS started_at,
			investigation_completed_at AS completed_at
		FROM payments p
		LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid
		LEFT JOIN merchants m ON p.merchant_id = m.uuid` + whereClause + partitionCondition

	sortBy := "investigation_started_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortDir := "DESC"
	if filter.Sort != "" && (strings.ToUpper(filter.Sort) == "ASC" || strings.ToUpper(filter.Sort) == "DESC") {
		sortDir = strings.ToUpper(filter.Sort)
	}
	query += fmt.Sprintf(" ORDER BY p.%s %s", sortBy, sortDir)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, offset)

	errG.Go(func() error {
		if err := r.db.SelectContext(ctx, &data, query, args...); err != nil {
			r.logger.Error(ctx, "error when getting investigated payments list", logger.Error(err))
			return err
		}
		return nil
	})

	var totalItems int64
	queryCount := `SELECT COUNT(*) as totalItems FROM payments p` + whereClause + partitionCondition

	errG.Go(func() error {
		_ = r.db.GetContext(ctx, &totalItems, queryCount, args...)
		return nil
	})

	if err := errG.Wait(); err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(filter.Limit)))
	meta := commonModel.Meta{
		Page:       int64(filter.Page),
		PerPage:    int64(filter.Limit),
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	return &commonModel.PaginationResponse{
		Data: data,
		Meta: meta,
	}, nil
}

func buildInvestigationCondition(filter *paymentModel.GetInvestigatedPaymentsFilterRequest) (conditions []string, args []any) {
	if filter.PaymentReferenceID != "" {
		conditions = append(conditions, "p.reference_id = ?")
		args = append(args, filter.PaymentReferenceID)
	}

	if filter.MerchantID != "" {
		conditions = append(conditions, "p.merchant_id = ?")
		args = append(args, filter.MerchantID)
	}

	if filter.FromDate != nil {
		conditions = append(conditions, "p.investigation_started_at >= ?")
		args = append(args, filter.FromDate)
	}

	if filter.ToDate != nil {
		conditions = append(conditions, "p.investigation_started_at <= ?")
		args = append(args, filter.ToDate)
	}

	if filter.PaymentMethod != "" {
		conditions = append(conditions, "pm.type = ?")
		args = append(args, filter.PaymentMethod)
	}

	if filter.Channel != "" {
		conditions = append(conditions, "pm.name = ?")
		args = append(args, filter.Channel)
	}

	if filter.InvestigationStatus != "" {
		conditions = append(conditions, "p.reason_type = ?")
		args = append(args, filter.InvestigationStatus)
	} else {
		conditions = append(conditions, "p.reason_type IN ('INVESTIGATION_IN_PROCESS', 'INVESTIGATION_SUCCESS', 'INVESTIGATION_FAILED')")
	}

	return conditions, args
}

func (r *PaymentRepository) UpdateInvestigationStatus(ctx context.Context, request paymentModel.UpdateInvestigationStatusRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/UpdateInvestigationStatus")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `UPDATE payments
		SET
			reason_type = ?, reason_description = ?, investigation_completed_at = ?, updated_at = ?
		WHERE uuid = ?;`

	_, err := r.db.ExecContext(
		ctx, query, request.Status, request.Notes, request.CompletedAt, time.Now().UTC(), request.PaymentID,
	)
	if err != nil {
		r.logger.Error(ctx, "error when updating investigation status", logger.Error(err))
		return err
	}
	return nil
}

func (r *PaymentRepository) CalculateInvestigationMonthlyReconciliation(ctx context.Context, request paymentModel.MonthlyReconciliationRequest) ([]paymentModel.CalculateInvestigationMonthlyReconciliation, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/CalculateInvestigationMonthlyReconciliation")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payments,account_transactions,merchants")

	rawQuery := `SELECT
		merchant_id, payment_ids, payment_count, gross_amount, fee_amount, (gross_amount - fee_amount) AS net_amount, 
		IFNULL(m.transaction_configs->>'$.paymentInvestigation.pivotPercentageLoss', 0) AS platform_loss_percentage,
		IFNULL(m.transaction_configs->>'$.paymentInvestigation.pivotMaxLoss', 0) AS platform_max_loss
	FROM (
		SELECT
			p.merchant_id, JSON_ARRAYAGG(p.uuid) AS payment_ids, COUNT(p.uuid) AS payment_count, SUM(p.total_amount) AS gross_amount, SUM(at.debit) AS fee_amount
		FROM payments p
		LEFT JOIN account_transactions at ON at.reference_id = p.uuid AND at.type = 'FEE' AND at.created_at >= DATE_SUB(NOW(), INTERVAL 3 MONTH)
		WHERE
			p.reason_type = ?
			AND (p.investigation_completed_at BETWEEN ? AND ?)
			AND p.created_at >= DATE_SUB(NOW(), INTERVAL 3 MONTH)
			AND p.metadata->>'$.investigationPoP.reconciledAt' IS NULL
		GROUP BY merchant_id
	) foo
	JOIN merchants m ON m.uuid = foo.merchant_id;`

	result := []paymentModel.CalculateInvestigationMonthlyReconciliation{}

	if err := r.db.SelectContext(ctx, &result, rawQuery, constant.InvestigationStatusFailed, request.StartDate, request.EndDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	for i, data := range result {

		data.PlatformLossAmount = data.NetAmount * (data.PlatformLossPercentage / 100)
		if data.PlatformLossAmount > 0.00 {
			data.PlatformLossAmount = decimal.NewFromFloat(data.PlatformLossAmount).Round(0).InexactFloat64()
		}
		if data.PlatformLossAmount > data.PlatformMaxLoss {
			data.PlatformLossAmount = data.PlatformMaxLoss
		}

		_ = data.RawPaymentIDs.Unmarshal(&result[i].PaymentIDs)
		result[i].PlatformLossAmount = data.PlatformLossAmount
		result[i].MerchantLossAmount = data.NetAmount - data.PlatformLossAmount
	}

	return result, nil
}

func (r *PaymentRepository) InsertInvestigationMonthlyReconciliation(ctx context.Context, data paymentModel.PaymentInvestigationMonthlyReconciliation) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/InsertInvestigationMonthlyReconciliation")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payment_investigation_monthly_reconciliations")

	rawQuery := `INSERT INTO payment_investigation_monthly_reconciliations(
		uuid, date, merchant_id, payment_ids, payment_count, gross_amount, fee_amount, net_amount, platform_loss_percentage, platform_max_loss, platform_loss_amount, merchant_loss_amount, created_at
	) VALUES(
		:uuid, :date, :merchant_id, :payment_ids, :payment_count, :gross_amount, :fee_amount, :net_amount, :platform_loss_percentage, :platform_max_loss, :platform_loss_amount, :merchant_loss_amount, :created_at 
	);`

	if _, err := r.db.NamedExecContext(ctx, rawQuery, data); err != nil {
		return err
	}
	return nil
}

func (r *PaymentRepository) UpdatePaymentInvestigationReconciliation(ctx context.Context, data paymentModel.PaymentInvestigationMonthlyReconciliation) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/UpdatePaymentInvestigationReconciliation")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `UPDATE
		payments
	SET
		metadata = JSON_SET(metadata, '$.investigationPoP.reconcileId', ?, '$.investigationPoP.reconciledAt', ?)
	WHERE
		uuid IN (?) AND created_at >= DATE_SUB(NOW(), INTERVAL 3 MONTH);`

	query, args, _ := sqlx.In(rawQuery, data.UUID, data.Date.Format(time.RFC3339), data.PaymentIDs)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return err
	}
	return nil
}
