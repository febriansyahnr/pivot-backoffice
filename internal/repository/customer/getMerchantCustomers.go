package customerRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *CustomerRepository) GetMerchantCustomersByID(ctx context.Context, merchantId string, customerIds []string) ([]*customerModel.Customer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/customer/GetMerchantCustomersByID")
	defer segment.End()

	var data []*customerModel.CustomerDBModel
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	customerIdParam := strings.Join(customerIds, "','")
	customerIdParam = fmt.Sprintf("'%s'", customerIdParam)

	query := `
			SELECT 
				uuid, 
				merchant_id, 
				email,
				phone_country_code,
				phone_number, 
				created_at, 
				updated_at, 
				deleted_at,
				first_name,
				last_name,
				business_name,
				metadata,
				city, 
				country,
				address_line1,
				address_line2,
				postal_code,
				state
			FROM ` + tableName + `
			WHERE 
				merchant_id = ?
				AND uuid IN (%s)
				AND deleted_at IS NULL `
	query = fmt.Sprintf(query, customerIdParam)
	if err := r.db.SelectContext(ctx, &data, query, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error(ctx, "merchant customers not found", logger.Error(err), logger.Any("merchantId", merchantId), logger.Any("customerIds", customerIds))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchant customers", logger.Error(err), logger.Any("merchantId", merchantId), logger.Any("customerIds", customerIds))
		return nil, err
	}

	var customers []*customerModel.Customer
	for _, item := range data {
		customers = append(customers, item.ToCustomerModel())
	}
	return customers, nil

}
