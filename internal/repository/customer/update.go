package customerRepository

import (
	"context"
	"errors"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *CustomerRepository) Update(ctx context.Context, customer *customerModel.Customer) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	query := `
			UPDATE ` + tableName + `
			SET
				merchant_id = :merchant_id, 
				email = :email,
				phone_country_code = :phone_country_code,
				phone_number = :phone_number, 
				updated_at = :updated_at, 
				first_name = :first_name,
				last_name = :last_name,
				business_name = :business_name,
				city = :city, 
				country = :country,
				address_line1 = :address_line1,
				address_line2 = :address_line2,
				postal_code = :postal_code,
				state = :state,
				metadata = :metadata,
				is_blocked = :is_blocked,
				block_reason = :block_reason
			WHERE
				uuid = :uuid;
			`

	affected, err := r.db.NamedExecContext(ctx, query, *customer.ToCustomerDBModel())
	if err != nil {
		r.logger.Error(ctx, "error when updating customer", logger.Error(err))
		return err
	}

	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when updating customer", logger.Error(err))
		return err
	}

	return nil
}
