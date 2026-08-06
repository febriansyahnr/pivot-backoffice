package customerRepository

import (
	"context"
	"errors"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *CustomerRepository) Create(ctx context.Context, customer *customerModel.Customer) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	query := `
			INSERT INTO ` + tableName + `(
				uuid, 
				merchant_id, 
				email,
				phone_country_code,
				phone_number, 
				first_name,
				last_name,
				business_name,
				city, 
				country,
				address_line1,
				address_line2,
				postal_code,
				state,
				metadata,
				is_blocked,
				block_reason,
				created_at, 
				updated_at
			) VALUES (
				:uuid, 
				:merchant_id, 
				:email,
				:phone_country_code,
				:phone_number, 
				:first_name,
				:last_name,
				:business_name,
				:city, 
				:country,
				:address_line1,
				:address_line2,
				:postal_code,
				:state,
				:metadata,
				:is_blocked,
				:block_reason,
				:created_at, 
				:updated_at
			)`

	affected, err := r.db.NamedExecContext(ctx, query, *customer.ToCustomerDBModel())
	if err != nil {
		r.logger.Error(ctx, "error when inserting customer", logger.Error(err))
		return err
	}

	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting customer", logger.Error(err))
		return err
	}

	return nil
}
