package bankAccountRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
)

func (r *bankAccountRepository) GetBankAccountValidation(ctx context.Context, merchantId, bankCode, accountNo string) (result *bankAccount.BankAccountResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/bankAccount/GetBankAccountValidation")
	defer segment.End()

	ctx, result = context.WithValue(ctx, constant.CtxSQLTableNameKey, tableName), &bankAccount.BankAccountResponse{}

	rawQuery := `SELECT 
			beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no, beneficiary_account_name
		FROM bank_accounts 
		WHERE merchant_id = ? AND beneficiary_bank_code = ? AND beneficiary_account_no = ? AND deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, result, rawQuery, merchantId, bankCode, accountNo); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}

// GetByMerchantID is a function to get bank account by merchant id
func (r *bankAccountRepository) GetByMerchantID(ctx context.Context, merchantId string) (result *bankAccount.BankAccount, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/bankAccount/GetByMerchantID")
	defer segment.End()

	ctx, result = context.WithValue(ctx, constant.CtxSQLTableNameKey, tableName), &bankAccount.BankAccount{}

	rawQuery := `SELECT 
			id, merchant_id, beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no,
			beneficiary_account_name, created_by, created_at, updated_by, updated_at
		FROM bank_accounts 
		WHERE merchant_id = ? AND deleted_at IS NULL;`
	if err = r.db.GetContext(ctx, result, rawQuery, merchantId); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}

func (r *bankAccountRepository) BankAccountHasBeenPrepared(ctx context.Context, merchantId string) (result bool, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/bankAccount/BankAccountHasBeenPrepared")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxSQLTableNameKey, tableName)

	rawQuery := `SELECT COUNT(id) >= 1 FROM bank_accounts WHERE merchant_id = ? AND deleted_at IS NULL;`

	err = r.db.GetContext(ctx, &result, rawQuery, merchantId)
	return
}
