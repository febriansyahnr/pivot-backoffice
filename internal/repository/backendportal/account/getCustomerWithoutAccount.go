package account_repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/customer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *AccountRepository) GetCustomersWithoutAccount(ctx context.Context, request account_model.GetEntityWithoutAccountRequest) ([]*customerModel.Customer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/GetCustomersWithoutAccount")
	defer segment.End()

	var (
		data []*customerModel.Customer
	)

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, TableName)
	query := `
		SELECT 
			c.uuid, c.merchant_id, c.full_name, c.email, c.phone_number, c.created_at, c.updated_at, c.deleted_at
		FROM
			` + CustomerTableName + ` as c
		LEFT JOIN ` + TableName + ` 
			as a ON c.uuid = a.reference_id 
			AND a.name = ? 
			AND a.user_type = ?
		WHERE
			(c.merchant_id = ?) AND a.uuid IS NULL AND c.deleted_at IS NULL
		LIMIT ? 
	`
	err := r.db.SelectContext(ctx, &data, query, request.Usecase, constant.UserTypeCustomer, request.MerchantID, request.Limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug(ctx, "customers without account is not found", logger.Any("request", request))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding customers without account", logger.Error(err), logger.Any("request", request))
		return nil, err

	}

	return data, nil
}
