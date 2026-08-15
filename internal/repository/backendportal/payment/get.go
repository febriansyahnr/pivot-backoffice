package paymentRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PaymentRepository) GetPaymentById(ctx context.Context, id string) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetPaymentById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments")

	var paymentWithMethodDTO paymentModel.PaymentWithPaymentMethodDTO
	query := `
			SELECT 
				p.uuid, 
				p.reference_id,
				p.merchant_id, 
				p.customer_id, 
				p.payment_method_id, 
				p.processor_reference_number, p.recurring_contract_id,
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
				p.created_by,
				p.expired_at,
				p.updated_at, 
				p.deleted_at,
				p.created_from,
				p.reason_type,
				p.reason_description,
				pm.type as 'payment_method_type',
				pm.name as 'payment_method_name',
				pm.acquirer as 'payment_method_acquirer',
				pm.logo as 'payment_method_logo',
				pm.bank_name as 'payment_method_bank_name',
				CASE
					WHEN pm.type IN ('CARD', 'CREDIT_CARD') AND p.metadata->>'$.paymentMethodOptions.card.captureMethod' = 'MANUAL' THEN
						(
							SELECT JSON_ARRAYAGG(
								JSON_OBJECT(
									'id', c.id,
									'paymentId', c.payment_id,
									'processorReferenceId', c.processor_reference_id,
									'status', c.status,
									'releaseRemainingAmount', c.release_remaining_amount,
									'currency', c.currency,
									'amount', c.amount,
									'createdAt', DATE_FORMAT(c.created_at, '%Y-%m-%dT%H:%i:%sZ'),
									'updatedAt', DATE_FORMAT(c.updated_at, '%Y-%m-%dT%H:%i:%sZ')
								)
							)
							FROM payment_captures c
							WHERE c.payment_id = p.uuid
						)
				    ELSE NULL
				END AS payment_captures
			FROM payments p
			LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid
			WHERE p.uuid = ?
			ORDER BY p.created_at ASC
			LIMIT 1 `

	if err := r.db.GetContext(ctx, &paymentWithMethodDTO, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding payments by id=%s", id), logger.Error(err))
		return nil, err
	}

	var payment paymentModel.Payment
	payment.PaymentFromPaymentWithPaymentMethodDTO(&paymentWithMethodDTO)

	return &payment, nil
}

func (r *PaymentRepository) GetPaymentItemsByPaymentId(
	ctx context.Context, paymentID string) ([]*paymentModel.PaymentItem, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetPaymentItemsByPaymentId")
	defer segment.End()

	var (
		paymentItems []*paymentModel.PaymentItem
	)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payment_items")

	var paymentItemsDTO []*paymentModel.PaymentItemDTO
	query := `
			SELECT 
				pi.uuid, 
				pi.payment_id,
				pi.name,
				pi.description,
				pi.qty,
				pi.currency,
				pi.amount,
				pi.total_amount,
				pi.metadata,
				pi.created_at,
				pi.updated_at,
				pi.deleted_at
			FROM payment_items pi
			WHERE pi.payment_id = ?`

	if err := r.db.SelectContext(ctx, &paymentItemsDTO, query, paymentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, fmt.Sprintf("get payment items is not found by payment_id=%s", paymentID), logger.Error(err))
			return nil, errors.New("payment items is not found")
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding payment items by payment_id=%s", paymentID), logger.Error(err))
		return nil, err
	}

	for _, item := range paymentItemsDTO {
		paymentItems = append(paymentItems, item.ToPaymentItem())
	}

	return paymentItems, nil
}

func (r *PaymentRepository) GetActivePaymentByProcessorReferenceNumber(ctx context.Context, request *paymentModel.GetActivePaymentByProcessorReferenceNumberRequest) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetActivePaymentByProcessorReferenceNumber")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments")

	var paymentWithMethodDTO paymentModel.PaymentWithPaymentMethodDTO
	query := `
			SELECT 
				p.uuid, 
				p.merchant_id,
				p.reference_id, 
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
				p.deleted_at,
				pm.type as 'payment_method_type',
				pm.name as 'payment_method_name',
				pm.acquirer as 'payment_method_acquirer'
			FROM payments p
			LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid
			`

	var (
		whereParams   = []any{}
		whereClause   = []string{}
		orderByClause = ` p.created_at DESC LIMIT 1`
	)
	whereClause = append(whereClause, "p.status IN (?,?,?,?)")
	whereParams = append(whereParams, paymentConstant.PAYMENT_STATUS_PENDING, paymentConstant.UnifiedPaymentStatusWaitingForPayment, constant.UnifiedStaticPaymentStatusActive, paymentConstant.UnifiedPaymentStatusExpired)

	if request.ProcessorReferenceNumber != "" {
		whereParams = append(whereParams, request.ProcessorReferenceNumber)
		whereClause = append(whereClause, "p.processor_reference_number = ?")
	}
	if len(whereClause) > 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
	}
	if !request.Amount.IsZero() {
		orderByClause = `IF(p.amount = '` + request.Amount.String() + `', 1, 0) DESC, ` + orderByClause
	}

	query += " ORDER BY " + orderByClause

	if err := r.db.GetContext(ctx, &paymentWithMethodDTO, query, whereParams...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding active payment by processor_reference_number=%s", request.ProcessorReferenceNumber), logger.Error(err))
		return nil, err
	}

	var payment paymentModel.Payment
	payment.PaymentFromPaymentWithPaymentMethodDTO(&paymentWithMethodDTO)

	paymentMethodType := payment.PaymentMethod.Type
	paymentStatus := payment.Status

	// We need to include expired status in this payment type, because source of truth is from bank notification
	allowedExpiredPaymentMethod := []string{
		paymentConstant.PAYMENT_METHOD_QRIS,
		paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
	}

	if paymentStatus == paymentConstant.UnifiedPaymentStatusExpired &&
		!slices.Contains(allowedExpiredPaymentMethod, paymentMethodType) {
		return nil, nil
	}

	return &payment, nil
}

func (r *PaymentRepository) GetPaymentByMerchantAndReferenceId(ctx context.Context, merchantId, referenceId string) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetPaymentByReferenceId")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments")

	var paymentWithMethodDTO paymentModel.PaymentWithPaymentMethodDTO
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
				p.deleted_at,
				pm.type as 'payment_method_type',
				pm.name as 'payment_method_name',
				pm.acquirer as 'payment_method_acquirer',
				p.metadata->>'$.snapCore.uuid' AS payment_snap_core_id
			FROM payments p
			LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid
			WHERE p.merchant_id = ?
			AND p.reference_id = ?
			ORDER BY p.created_at ASC
			LIMIT 1 `

	if err := r.db.GetContext(ctx, &paymentWithMethodDTO, query, merchantId, referenceId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, fmt.Sprintf("get payments is not found by merchant id = %s and reference id = %s", merchantId, referenceId), logger.Error(err))
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding payments by merchant id = %s and reference id = %s", merchantId, referenceId), logger.Error(err))
		return nil, err
	}

	var payment paymentModel.Payment
	payment.PaymentFromPaymentWithPaymentMethodDTO(&paymentWithMethodDTO)

	return &payment, nil
}

func (r *PaymentRepository) GetPaymentByIdAndMerchantId(ctx context.Context, id, merchantId string) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetPaymentByIdAndMerchantId")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments")

	var paymentWithMethodDTO paymentModel.PaymentWithPaymentMethodDTO
	query := `
			SELECT 
				p.uuid, 
				p.reference_id,
				p.merchant_id, 
				p.customer_id, 
				p.payment_method_id, 
				p.recurring_contract_id,
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
				p.created_from,
				p.expired_at,
				p.updated_at, 
				p.deleted_at,
				p.investigation_started_at,
				pm.type as 'payment_method_type',
				pm.name as 'payment_method_name',
				pm.acquirer as 'payment_method_acquirer',
				pm.bank_name as 'payment_method_bank_name'
			FROM payments p
			LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid
			WHERE p.uuid = ?
			AND p.merchant_id = ?
			ORDER BY p.created_at ASC
			LIMIT 1 `

	if err := r.db.GetContext(ctx, &paymentWithMethodDTO, query, id, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, fmt.Sprintf("get payments is not found by id = %s and merchant id = %s", id, merchantId), logger.Error(err))
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding payments by id = %s and merchant id = %s", id, merchantId), logger.Error(err))
		return nil, err
	}

	var payment paymentModel.Payment
	payment.PaymentFromPaymentWithPaymentMethodDTO(&paymentWithMethodDTO)

	return &payment, nil
}

func (r *PaymentRepository) GetPaymentQrStaticByMerchantId(ctx context.Context, merchantId string, subMerchantId string, paymentMethodId string) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetPaymentQrStaticByMerchantId")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments")

	var paymentDTO paymentModel.PaymentDTO
	// Base query
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
			FROM payments p
			WHERE p.merchant_id = ?
			AND p.payment_method_id = ?
			AND p.metadata->>'$.qrType' = 'STATIC'
			AND p.metadata->>'$.subMerchantId' = ?
			ORDER BY p.created_at ASC
			LIMIT 1
	`

	if err := r.db.GetContext(ctx, &paymentDTO, query, merchantId, paymentMethodId, subMerchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, fmt.Sprintf("get payment qr static is not found by merchant id = %s, submerchant id = %s, and payment method id = %s", merchantId, subMerchantId, paymentMethodId), logger.Error(err))
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error when finding qr static is not found by merchant id = %s, submerchant id = %s, and payment method id = %s", merchantId, subMerchantId, paymentMethodId), logger.Error(err))
		return nil, err
	}

	var payment paymentModel.Payment
	payment.PaymentFromDTO(&paymentDTO)

	return &payment, nil
}

// GetPaymentReceiptData fetches payment, payment method, and merchant data in a single query
// for generating payment receipts. Returns nil if not found or not accessible.
func (r *PaymentRepository) GetPaymentReceiptData(ctx context.Context, paymentID, referenceID, merchantID string) (*paymentModel.PaymentReceiptDTO, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/GetPaymentReceiptData")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments")

	query := `
		SELECT
			p.uuid,
			p.reference_id,
			p.merchant_id,
			p.total_amount,
			p.status,
			p.created_at,
			m.name as merchant_name,
			pm.type as payment_method_type
		FROM payments p
		LEFT JOIN merchants m ON p.merchant_id = m.uuid
		LEFT JOIN payment_methods pm ON p.payment_method_id = pm.uuid
		WHERE 1=1`

	var args []interface{}

	if paymentID != "" {
		query += " AND p.uuid = ?"
		args = append(args, paymentID)
	}
	if referenceID != "" {
		query += " AND p.reference_id = ?"
		args = append(args, referenceID)
	}
	if merchantID != "" {
		query += " AND p.merchant_id = ?"
		args = append(args, merchantID)
	}

	query += " ORDER BY p.created_at DESC LIMIT 1"

	var result paymentModel.PaymentReceiptDTO
	if err := r.db.GetContext(ctx, &result, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, fmt.Sprintf("error getting payment receipt data: paymentID=%s, referenceID=%s, merchantID=%s", paymentID, referenceID, merchantID), logger.Error(err))
		return nil, err
	}

	return &result, nil
}
