package beneficiaryAccountRepository

import (
	"context"

	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *BeneficiaryAccountRepository) Upsert(
	ctx context.Context, account *beneficiaryAccountModel.BeneficiaryAccount) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/beneficiaryAccount/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "beneficiary_account")

	query := `
		INSERT INTO beneficiary_accounts (uuid, merchant_id, beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no, beneficiary_account_name, metadata, created_at, updated_at, deleted_at)
		VALUES (:uuid, :merchant_id, :beneficiary_bank_code, :beneficiary_bank_name, :beneficiary_account_no, :beneficiary_account_name, :metadata, :created_at, :updated_at, :deleted_at)
		ON DUPLICATE KEY UPDATE
		    merchant_id = VALUES(merchant_id),
		    beneficiary_bank_code = VALUES(beneficiary_bank_code),
		    beneficiary_bank_name = VALUES(beneficiary_bank_name),
		    beneficiary_account_no = VALUES(beneficiary_account_no),
		    beneficiary_account_name = VALUES(beneficiary_account_name),
		    metadata = VALUES(metadata),
		    updated_at = VALUES(updated_at),
		    deleted_at = VALUES(deleted_at);
	`

	_, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		r.logger.Error(ctx, "error when upsert beneficiary account", logger.Error(err))
		return err
	}

	return nil
}
