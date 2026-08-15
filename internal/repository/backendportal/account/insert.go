package account_repository

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

// Create implements repository.IBalanceRepository.
func (r *AccountRepository) Create(
	ctx context.Context,
	account *account_model.Account,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "accounts")

	query := `
		INSERT INTO accounts (
			uuid, reference_id, name, eod_balance, 
			currency, last_update_balance_at, type, user_type, created_at, updated_at
		) VALUES (
            :uuid, :reference_id, :name, :eod_balance, 
            :currency, :last_update_balance_at, :type, :user_type, :created_at, :updated_at
        )`

	affected, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		r.logger.Error(ctx, "error when inserting account", logger.Error(err))
		return err
	}

	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting account", logger.Error(err))
		return err
	}

	return nil
}
