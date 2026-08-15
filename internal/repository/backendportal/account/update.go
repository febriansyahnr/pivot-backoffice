package account_repository

import (
	"context"

	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountRepository) UpdateAccount(
	ctx context.Context,
	account *account_model.Account,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/UpdateAccount")
	defer segment.End()

	query := `
		UPDATE accounts
		SET
			eod_balance = ?,
			last_update_balance_at = ?,
			pending_transaction_cutoff_at = ?
		WHERE uuid = ?`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "accounts")

	_, err := r.db.ExecContext(
		ctx,
		query,
		account.EODBalance,
		account.LastUpdateBalanceAt,
		account.PendingTransactionCutoffAt,
		account.UUID,
	)
	if err != nil {
		r.logger.Error(ctx, "error when updating account", logger.Error(err))
		return err
	}

	return nil
}

func (r *AccountRepository) UpdateHoldedBalance(
	ctx context.Context,
	account *account_model.Account,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/account/UpdateHoldedBalance")
	defer segment.End()

	query := `
		UPDATE accounts
		SET
			holded_balance = ?
		WHERE uuid = ?`

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "accounts")

	_, err := r.db.ExecContext(
		ctx,
		query,
		account.HoldedBalance,
		account.UUID,
	)
	if err != nil {
		r.logger.Error(ctx, "error when updating holded balance", logger.Error(err))
		return err
	}

	return nil
}
