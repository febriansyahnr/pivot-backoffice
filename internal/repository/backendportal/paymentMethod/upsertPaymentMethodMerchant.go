package paymentMethodRepository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *PaymentMethodRepository) UpsertPaymentMethodMerchantByIdAndMerchant(ctx context.Context, paymentMethodMerchant *paymentModel.PaymentMethodWithPivot) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/paymentMethod/UpsertActivationByIdAndMerchant")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	var (
		args []interface{}
	)

	query := `INSERT INTO payment_method_merchant (uuid, merchant_id, payment_method_id, is_active, activation_status, channel_type, config, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		is_active = VALUES(is_active),
		activation_status = VALUES(activation_status),
		channel_type = VALUES(channel_type),
		config = VALUES(config),
		updated_at = VALUES(updated_at);`

	args = append(args,
		uuid.NewString(),
		paymentMethodMerchant.MerchantID,
		paymentMethodMerchant.UUID,
		paymentMethodMerchant.IsActive,
		paymentMethodMerchant.ActivationStatus,
		paymentMethodMerchant.ChannelType,
		paymentMethodMerchant.MerchantConfig,
		time.Now().UTC(),
		time.Now().UTC(),
	)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error(ctx, "error when upsert payment method merchant", logger.Error(err))
		return err
	}

	return nil
}
