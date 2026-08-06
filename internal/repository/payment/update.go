package paymentRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (r *PaymentRepository) UpdatePayment(
	ctx context.Context, id string, amount, totalAmount decimal.Decimal, metadata string, customerId string, expiredAt time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/UpdatePayment")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE payments
		SET
		    amount = ?, 
			total_amount = ?, 
			metadata = ?,
			customer_id = ?, 
			updated_at = CURRENT_TIMESTAMP(),
			expired_at = ?
		WHERE uuid = ?`

	_, err := r.db.ExecContextReturnLastId(ctx, query, amount, totalAmount, metadata, customerId, expiredAt, id)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating payment", logger.Error(err))
		return err
	}

	return nil
}

func (r *PaymentRepository) UpdatePaymentStatus(
	ctx context.Context, id string, merchantId string, status string, updatedAt time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/UpdatePaymentStatus")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE payments
		SET
			status = ?, 
			updated_at = ?
		WHERE uuid = ?
		AND merchant_id = ?`

	_, err := r.db.ExecContext(
		ctx, query, status, updatedAt, id, merchantId,
	)
	if err != nil {
		r.logger.Error(ctx, "error when updating payment", logger.Error(err))
		return err
	}

	return nil
}

func (r *PaymentRepository) UpdatePaymentItemsFromPaymentResponseItem(ctx context.Context, paymentID string, paymentRespItems []paymentModel.PaymentResponseItem) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/UpdatePaymentItemsFromPaymentResponseItem")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "payment_items")

	ctx, err := r.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	// Delete all payment items by payment id
	query := `DELETE FROM payment_items WHERE payment_id = ?`
	_, err = r.db.ExecContext(ctx, query, paymentID)
	if err != nil {
		r.logger.Error(ctx, "error when deleting payment items by payment id", logger.Error(err))
		if rollbackErr := r.RollbackTransaction(ctx); rollbackErr != nil {
			r.logger.Error(ctx, "failed to rollback transaction", logger.Error(rollbackErr))
		}
		return err
	}

	for _, item := range paymentRespItems {
		if item.ItemID == "" {
			item.ItemID = uuid.New().String()
		}

		paymentDTO := &paymentModel.PaymentItemDTO{
			UUID:        item.ItemID,
			PaymentID:   paymentID,
			Name:        item.Name,
			Description: item.Description,
			Amount:      item.Amount.Value,
			Currency:    item.Amount.Currency,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		err = r.CreatePaymentItem(ctx, paymentDTO)

		if err != nil {
			r.logger.Error(ctx, "error when inserting payment items", logger.Error(err))
			if rollbackErr := r.RollbackTransaction(ctx); rollbackErr != nil {
				r.logger.Error(ctx, "failed to rollback transaction", logger.Error(rollbackErr))
			}
			return err
		}
	}

	if err = r.CommitTransaction(ctx); err != nil {
		return err
	}

	return nil
}

func (r *PaymentRepository) UpdatePaymentData(
	ctx context.Context, payment *paymentModel.PaymentDTO) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/UpdatePaymentData")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE payments
		SET
			processor_reference_number = ?,
		    amount = ?, 
			fee = ?,
			discount = ?,
			total_amount = ?, 
			customer_id = ?, 
			payment_method_id = ?,
			metadata = ?,
			status = ?,
			expired_at = ?,
			updated_at = ?
		WHERE uuid = ?
		AND merchant_id = ?`

	_, err := r.db.ExecContextReturnLastId(ctx, query,
		payment.ProcessorReferenceNumber,
		payment.Amount,
		payment.Fee,
		payment.Discount,
		payment.TotalAmount,
		payment.CustomerID,
		payment.PaymentMethodID,
		payment.Metadata,
		payment.Status,
		payment.ExpiredAt,
		payment.UpdatedAt,
		payment.UUID,
		payment.MerchantID)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating payment", logger.Error(err))
		return err
	}

	return nil
}

func (r *PaymentRepository) UpdatePaymentMetadataById(ctx context.Context, id string, metadata paymentModel.UpdatePaymentMetadataRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/UpdatePaymentMetadataById")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	jsonFieldSet, fields, values := []string{}, []string{"updated_at = ?"}, []any{time.Now().UTC()}

	if metadata.FeeDetail != nil {
		raw, _ := json.Marshal(metadata.FeeDetail)
		jsonFieldSet = append(jsonFieldSet, fmt.Sprintf("'$.feeDetail', CAST('%s' AS JSON)", string(raw)))
	}

	if metadata.FeeOnBehalf != nil {
		raw, _ := json.Marshal(metadata.FeeOnBehalf)
		jsonFieldSet = append(jsonFieldSet, fmt.Sprintf("'$.feeOnBehalf', CAST('%s' AS JSON)", string(raw)))
	}

	if metadata.SummaryTransaction != nil {
		raw, _ := json.Marshal(metadata.SummaryTransaction)
		jsonFieldSet = append(jsonFieldSet, fmt.Sprintf("'$.summaryTransaction', CAST('%s' AS JSON)", string(raw)))
	}

	if metadata.FingerprintID != nil {
		jsonFieldSet = append(jsonFieldSet, fmt.Sprintf("'$.fingerprintId', '%s'", metadata.FingerprintID))
	}

	if metadata.IsSnap != nil {
		jsonFieldSet = append(jsonFieldSet, fmt.Sprintf("'$.isSnap', %v", metadata.IsSnap))
	}

	if len(jsonFieldSet) > 0 {
		fields = append(fields, fmt.Sprintf("metadata = JSON_SET(metadata, %s)", strings.Join(jsonFieldSet, ", ")))
	}

	values = append(values, id)
	rawQuery := "UPDATE payments SET " + strings.Join(fields, ", ") + " WHERE uuid = ?;"

	_, err := r.db.ExecContext(ctx, rawQuery, values...)
	return err
}

func (r *PaymentRepository) UpdatePaymentForInvestigation(
	ctx context.Context,
	request paymentModel.UpdatePaymentForInvestigationRequest,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/UpdatePaymentForInvestigation")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawJSON, err := json.Marshal(request.InvestigationMetadata)
	if err != nil {
		r.logger.Error(ctx, "error marshaling investigation metadata", logger.Error(err))
		return err
	}

	query := `UPDATE payments
		SET
			reason_type = ?,
			metadata = JSON_SET(COALESCE(metadata, '{}'), '$.investigationPoP', CAST(? AS JSON)),
			investigation_started_at = ?, updated_at = ?
		WHERE 
			uuid = ? AND merchant_id = ?;`

	_, err = r.db.ExecContext(
		ctx, query, request.ReasonType, string(rawJSON), request.StartedAt, time.Now().UTC(), request.PaymentID, request.MerchantID,
	)
	if err != nil {
		r.logger.Error(ctx, "error when updating payment for investigation", logger.Error(err))
		return err
	}

	return nil
}

func (r *PaymentRepository) UpdatePaymentStatusWithReason(ctx context.Context, id string, request paymentModel.UpdatePaymentStatusWithReasonRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payment/UpdatePaymentStatusWithReason")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `UPDATE 
		payments 
	SET 
		status = ?, reason_type = ?, reason_description = ?, updated_at = ? 
	WHERE uuid = ?;`

	args := []any{request.Status, request.ReasonType, request.ReasonDescription, time.Now().UTC(), id}

	if _, err := r.db.ExecContext(ctx, rawQuery, args...); err != nil {
		return err
	}
	return nil
}
