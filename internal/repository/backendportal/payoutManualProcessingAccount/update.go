package payoutManualProcessingAccount

import (
	"context"

	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payoutManualProcessingAccount"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *PayoutManualProcessingAccountRepository) Update(ctx context.Context, account *payoutManualProcessingAccountModel.PayoutManualProcessingAccount) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/payoutManualProcessingAccount/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE payout_manual_processing_accounts SET
			status = :status,
			updated_by = :updated_by
		WHERE uuid = :uuid`

	_, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		r.logger.Error(ctx, "error when updating payout manual processing account", logger.Error(err))
		return err
	}

	return nil
}
