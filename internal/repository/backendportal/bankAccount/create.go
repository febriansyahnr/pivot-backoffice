package bankAccountRepository

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/bankAccount"
)

func (r *bankAccountRepository) Create(ctx context.Context, req *model.BankAccount) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/bankAccount/CreateBankAccount")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxSQLTableNameKey, tableName)

	rawQuery := `INSERT INTO bank_accounts (
                           id, merchant_id, beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no,
                           beneficiary_account_name, created_by, created_at, updated_by, updated_at
			   ) VALUES (
			             :id, :merchant_id, :beneficiary_bank_code, :beneficiary_bank_name, :beneficiary_account_no,
			             :beneficiary_account_name, :created_by, :created_at, :updated_by, :updated_at
			    )`
	affected, err := r.db.NamedExecContext(ctx, rawQuery, req)
	if err != nil {
		r.logger.Error(ctx, "error when inserting bank account", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when inserting bank account", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}

	return nil
}
