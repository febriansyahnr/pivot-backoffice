package paymentRepository

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *PaymentRepository) CreatePayment(ctx context.Context, paymentDTO *paymentModel.PaymentDTO) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/CreatePayment")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payments")

	query := `
		INSERT INTO payments (uuid, reference_id, merchant_id, customer_id, payment_method_id, processor_reference_number, currency, amount, fee, discount, total_amount, status, type, metadata, payment_url, created_at, expired_at, updated_at, deleted_at, created_by, created_from, recurring_contract_id)
		VALUES (:uuid, :reference_id, :merchant_id, :customer_id, :payment_method_id, :processor_reference_number, :currency, :amount, :fee, :discount, :total_amount, :status, :type, :metadata, :payment_url, :created_at, :expired_at, :updated_at, :deleted_at, :created_by, :created_from, :recurring_contract_id)
	`

	affected, err := r.db.NamedExecContext(ctx, query, paymentDTO)
	if err != nil {
		r.logger.Error(ctx, "error when inserting payments", logger.Error(err))
		return err
	}
	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting payments", logger.Error(err))
		return err
	}

	return nil
}

func (r *PaymentRepository) CreatePaymentItem(ctx context.Context, paymentItemDTO *paymentModel.PaymentItemDTO) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/payment/CreatePaymentItem")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "payment_items")

	query := `
		INSERT INTO payment_items (uuid, payment_id, name, description, qty, currency, amount, total_amount, metadata, created_at, updated_at, deleted_at)
		VALUES (:uuid, :payment_id, :name, :description, :qty, :currency, :amount, :total_amount, :metadata, :created_at, :updated_at, :deleted_at)
	`
	affected, err := r.db.NamedExecContext(ctx, query, paymentItemDTO)
	if err != nil {
		r.logger.Error(ctx, "error when inserting payment_items", logger.Error(err))
		return err
	}
	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting payment_items", logger.Error(err))
		return err
	}

	return nil
}
