package beneficiaryAccountRepository

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/beneficiaryAccount"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *BeneficiaryAccountRepository) Create(
	ctx context.Context, account *beneficiaryAccountModel.BeneficiaryAccount) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/beneficiaryAccount/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "beneficiary_account")

	query := `
		INSERT INTO beneficiary_accounts (uuid, merchant_id, beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no, beneficiary_account_name, metadata, created_at, updated_at, deleted_at)
		VALUES (:uuid, :merchant_id, :beneficiary_bank_code, :beneficiary_bank_name, :beneficiary_account_no, :beneficiary_account_name, :metadata, :created_at, :updated_at, :deleted_at)
	`

	affected, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		r.logger.Error(ctx, "error when inserting beneficiary account", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when inserting beneficiary account", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}

	return nil
}
