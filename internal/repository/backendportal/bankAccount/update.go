package bankAccountRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/bankAccount"
)

func (r *bankAccountRepository) Update(ctx context.Context, bk *bankAccount.BankAccount) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/bankAccount/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxSQLTableNameKey, tableName)
	query := `UPDATE bank_accounts
					SET
					    beneficiary_account_name = :beneficiary_account_name,
					    beneficiary_account_no = :beneficiary_account_no,
					    beneficiary_bank_code = :beneficiary_bank_code,
					    beneficiary_bank_name = :beneficiary_bank_name,
						updated_at = :updated_at
					WHERE id = :id;`

	_, err = r.db.NamedExecContext(ctx, query, bk)
	return
}
