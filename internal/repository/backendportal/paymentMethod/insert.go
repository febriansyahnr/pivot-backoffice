package paymentMethodRepository

import (
	"context"

	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PaymentMethodRepository) CreatePaymentMethod(ctx context.Context, payload *paymentMethodModel.CreatePaymentMethodRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/paymentMethod/CreatePaymentMethod")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `INSERT INTO payment_methods 
		(uuid, type, sub_type, category, name, description, logo, acquirer, bank_name, activation_method, country_of_operation, supported_currency, processor, instructions)
		VALUES(
		 	:uuid, :type, :sub_type, :category, :name, :description, :logo, :acquirer, :bank_name, :activation_method, :country_of_operation, :supported_currency, :processor, :instructions)`

	_, err := r.db.NamedExecContext(ctx, query, payload)
	if err != nil {
		r.logger.Error(ctx, "error when insert new payment method", logger.Error(err))
		return err
	}

	return nil
}
