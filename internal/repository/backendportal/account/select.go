package account_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"
)

func (r *AccountRepository) GetByUUID(ctx context.Context, accountId uuid.UUID) (*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/GetByUUID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "accounts")

	var account account_model.Account
	query := `
			SELECT 
				uuid, 
				reference_id, 
				name, 
				eod_balance, 
				currency, 
				last_update_balance_at,
				type,
				user_type,
				created_at,
				updated_at 
			FROM accounts 
			WHERE uuid = ?`

	if err := r.db.GetContext(ctx, &account, query, accountId.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		r.logger.Error(ctx, "error when getting account by uuid", logger.Error(err))
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepository) GetByIDs(ctx context.Context, accountIds []string) ([]*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/GetByUUID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "accounts")

	var accounts []*account_model.Account
	query := `
			SELECT 
				uuid, 
				reference_id, 
				name, 
				eod_balance, 
				currency, 
				last_update_balance_at,
				type,
				user_type,
				created_at,
				updated_at 
			FROM accounts 
			WHERE ` + fmt.Sprintf("uuid IN ('%s')", strings.Join(accountIds, "','")) + ` AND deleted_at IS NULL`

	if err := r.db.SelectContext(ctx, &accounts, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when getting accounts by ids", logger.Error(err))
		return nil, err
	}

	return accounts, nil
}

func (r *AccountRepository) GetEntityAccounts(ctx context.Context, entityIDs []uuid.UUID, userType, name string) (map[uuid.UUID]*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/GetEntityAccounts")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "accounts")

	var accounts []*account_model.Account
	query := `
			SELECT 
				uuid, 
				reference_id, 
				name, 
				eod_balance, 
				currency, 
				last_update_balance_at,
				type,
				user_type,
				created_at,
				updated_at 
			FROM accounts 
			WHERE reference_id IN (?) and user_type = ? and name = ? and deleted_at is null`

	query, args, err := sqlx.In(query, entityIDs, userType, name)
	if err != nil {
		r.logger.Error(ctx, "error when prepare where IN query", logger.Error(err), logger.Any("accountIds", entityIDs), logger.Any("userType", userType), logger.Any("name", name))
		return nil, err
	}
	query = r.db.Rebind(query)

	if err := r.db.SelectContext(ctx, &accounts, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		r.logger.Error(ctx, "error when getting account by entity ids", logger.Error(err), logger.Any("query", query), logger.Any("accountIds", entityIDs), logger.Any("userType", userType), logger.Any("name", name))
		return nil, err
	}

	accountMap := map[uuid.UUID]*account_model.Account{}
	for _, account := range accounts {
		accountMap[account.ReferenceID] = account
	}

	return accountMap, nil
}

func (r *AccountRepository) FindMerchantAccountByName(ctx context.Context, merchantID uuid.UUID, name string) (*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/FindMerchantAccountByName")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "accounts")

	var account account_model.Account
	query := `
			SELECT 
				uuid, 
				reference_id, 
				name, 
				eod_balance, 
				holded_balance,
				currency, 
				last_update_balance_at,
				pending_transaction_cutoff_at,
				type,
				user_type,
				created_at,
				updated_at 
			FROM accounts 
			WHERE reference_id = ? AND name = ?
			ORDER BY created_at desc
			LIMIT 1;`

	if err := r.db.GetContext(ctx, &account, query, merchantID.String(), name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding merchant account by name", logger.Error(err))
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepository) FindAll(ctx context.Context) ([]*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/FindAll")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "accounts")

	var accounts []*account_model.Account
	query := `
			SELECT 
				uuid, 
				reference_id, 
				name, 
				eod_balance, 
				currency, 
				last_update_balance_at,
				pending_transaction_cutoff_at,
				type,
				user_type,
				created_at,
				updated_at 
			FROM accounts
			WHERE deleted_at IS NULL`

	if err := r.db.SelectContext(ctx, &accounts, query); err != nil {
		r.logger.Error(ctx, "error when finding all accounts", logger.Error(err))
		return nil, err
	}

	return accounts, nil
}
