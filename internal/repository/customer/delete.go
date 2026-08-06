package customerRepository

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *CustomerRepository) Delete(ctx context.Context, id, merchantId string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/Delete")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	now := time.Now().UTC()
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE `+tableName+` SET deleted_at = ?, updated_at = ? WHERE uuid = ? AND merchant_id = ?;`, now, now, id, merchantId,
	)

	if err != nil {
		r.logger.Error(ctx, "error when deleting customer", logger.Error(err))
		return err
	}
	return nil
}

// RemovePaymentMethodFromCustomerByIDAndTokenID function to move payment method from paymentMethods path to deletePaymentMethods path based on customer id and token id.
func (r *CustomerRepository) RemovePaymentMethodFromCustomerByIDAndTokenID(ctx context.Context, id, tokenId string, paymentMethods []*unifiedPaymentModel.CustomerPaymentMethodResponse) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/customer/RemovePaymentMethodFromCustomerByIDAndTokenID")
	defer segment.End()

	var deletePaymentMethod *unifiedPaymentModel.CustomerPaymentMethodResponse

	paymentMethods = slices.DeleteFunc(paymentMethods, func(method *unifiedPaymentModel.CustomerPaymentMethodResponse) bool {
		if method != nil && method.Token == tokenId && deletePaymentMethod == nil {
			deletePaymentMethod = method
			return true
		}
		return false
	})

	if deletePaymentMethod == nil {
		return constant.ErrDataNotFound
	}

	rawQuery := `UPDATE
		customers
	SET
		metadata = JSON_SET(metadata, '$.paymentMethods', CAST(? AS JSON)),
		metadata = CASE WHEN metadata->>'$.deletePaymentMethods' IS NULL
			THEN
				JSON_SET(metadata, '$.deletePaymentMethods', JSON_ARRAY(CAST(? AS JSON)))
			ELSE
				JSON_ARRAY_APPEND(metadata, '$.deletePaymentMethods', CAST(? AS JSON))
			END,
		updated_at = ?
	WHERE
		uuid = ?;`

	rawPaymentMethods, _ := json.Marshal(paymentMethods)
	rawDeletePaymentMethod, _ := json.Marshal(deletePaymentMethod)

	_, err := r.db.ExecContext(ctx, rawQuery, rawPaymentMethods, rawDeletePaymentMethod, rawDeletePaymentMethod, time.Now().UTC(), id)
	return err
}
