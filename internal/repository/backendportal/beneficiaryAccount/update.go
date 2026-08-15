package beneficiaryAccountRepository

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/beneficiaryAccount"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *BeneficiaryAccountRepository) Update(
	ctx context.Context, account *beneficiaryAccountModel.BeneficiaryAccount) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/beneficiaryAccount/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "beneficiary_accounts")

	query := `
		UPDATE beneficiary_accounts
		SET
			beneficiary_bank_code = :beneficiary_bank_code,
			beneficiary_bank_name = :beneficiary_bank_name,
			beneficiary_account_no = :beneficiary_account_no,
			beneficiary_account_name = :beneficiary_account_name,
			metadata = :metadata,
			updated_at = :updated_at
		WHERE
			uuid = :uuid
	`

	affected, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		r.logger.Error(ctx, "error when updating beneficiary account", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when updating beneficiary account", logger.Error(errors.New("no rows affected")))
		return constant.ErrNoRowsAffected
	}

	return nil
}
