package paymentRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *PaymentRepository) GetAutoSplitSubPayments(ctx context.Context, request *paymentModel.GetSubPaymentsRequest) ([]*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetAutoSplitSubPayments")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var (
		data = make([]*paymentModel.PaymentWithPaymentMethodDTO, 0)
	)

	query := `
			SELECT 
				p.uuid, 
				p.reference_id,
				p.merchant_id, 
				p.customer_id, 
				p.payment_method_id, 
				p.processor_reference_number, 
				p.currency, 
				p.amount, 
				p.fee, 
				p.discount, 
				p.total_amount, 
				p.status, 
				p.type,
				p.metadata, 
				p.payment_url,
				p.created_at, 
				p.expired_at,
				p.updated_at, 
				p.deleted_at
			FROM payments p`

	conditions := []string{}
	args := []interface{}{}

	// Query condition builder
	conditions = append(conditions, fmt.Sprintf("p.type = '%s'", constant.UnifiedPaymentTypeSubPayment))
	conditions = append(conditions, fmt.Sprintf("p.merchant_id = '%s'", request.MerchantID))
	conditions = append(conditions, fmt.Sprintf("p.reference_id = '%s'", request.ReferenceID))

	if request.Status != "" {
		conditions = append(conditions, "p.status = ?")
		args = append(args, strings.ToUpper(request.Status))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	err := r.db.SelectContext(ctx, &data, query, args...)
	if err != nil {
		r.logger.Error(ctx, "error when get payments list", logger.Error(err))
		return nil, err
	}

	return BuildRespData(data), nil
}

func (r *PaymentRepository) HardDeleteAutoSplitSubPayments(ctx context.Context, merchantID, referenceID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/HardDeleteAutoSplitSubPayments")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payments,account_transactions")

	rawQuery := `DELETE p, at FROM payments p JOIN account_transactions at ON at.reference_id = p.uuid WHERE p.merchant_id = ? AND p.reference_id = ? AND p.type = ?;`

	if _, err := r.db.ExecContext(ctx, rawQuery, merchantID, referenceID, constant.UnifiedPaymentTypeSubPayment); err != nil {
		return err
	}
	return nil
}

func (r *PaymentRepository) GetSummaryAutoSplitPayment(ctx context.Context, request *paymentModel.GetAutoSplitPaymentSummaryRequest) (*paymentModel.AutoSplitPaymentSummary, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetSummaryAutoSplitPayment")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payments,account_transactions")

	// should include parent data into dataset
	rawQuery := `SELECT
		p.reference_id AS parent_payment_id,
		COUNT(p.uuid) as total_charge,
		SUM(IF(p.type = 'SUB_PAYMENT' AND p.status IN ('PROCESSING', 'REQUIRE_ACTION'), 1, 0)) AS total_in_progress_charge,
		SUM(IF(p.type = 'SUB_PAYMENT' AND p.status = 'CANCELLED', 1, 0)) AS total_failed_charge,
		SUM(IF(p.type = 'SUB_PAYMENT' AND p.status = 'EXPIRED', 1, 0)) AS total_expired_charge,
		SUM(IF(p.type = 'SUB_PAYMENT' AND p.status = 'PAID', 1, 0)) AS total_success_charge,
		SUM(IF(p.type = 'SUB_PAYMENT' AND p.status IN ('PROCESSING', 'REQUIRE_ACTION'), p.total_amount, 0)) AS total_in_progress_charge_amount,
		SUM(IF(p.type = 'SUB_PAYMENT' AND p.status IN ('CANCELLED', 'EXPIRED'), p.total_amount, 0)) AS total_failed_charge_amount,
		SUM(IF(p.type = 'SUB_PAYMENT' AND p.status = 'PAID', p.total_amount, 0)) AS total_success_charge_amount
	FROM payments p
	WHERE 
		p.merchant_id = ? AND p.reference_id = ? AND p.created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
	GROUP BY parent_payment_id;`

	result := &paymentModel.AutoSplitPaymentSummary{}
	err := r.db.GetContext(ctx, result, rawQuery, request.MerchantID, request.ReferenceID, request.MaxDateCreation)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return result, nil
}
