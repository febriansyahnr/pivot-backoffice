package payoutManualProcessingAccount

import (
	"context"

	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payoutManualProcessingAccount"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PayoutManualProcessingAccountRepository) Create(ctx context.Context, account *payoutManualProcessingAccountModel.PayoutManualProcessingAccount) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payoutManualProcessingAccount/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		INSERT INTO payout_manual_processing_accounts (
			uuid, merchant_id, bank_code, account_number, status, updated_by
		) VALUES (
			:uuid, :merchant_id, :bank_code, :account_number, :status, :updated_by
		)`

	_, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		r.logger.Error(ctx, "error when creating payout manual processing account", logger.Error(err))
		return err
	}

	return nil
}
