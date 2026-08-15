package bankAccountRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
)

func (r *bankAccountRepository) GetListBankAccount(ctx context.Context, merchantId string) (result []bankAccount.BankAccountResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/bankAccount/GetListBankAccount")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT 
			beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no, beneficiary_account_name
		FROM bank_accounts 
		WHERE merchant_id = ? AND deleted_at IS NULL ORDER BY beneficiary_bank_code, beneficiary_account_name;`

	if err = r.db.SelectContext(ctx, &result, rawQuery, merchantId); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}
