package merchant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func (r *MerchantRepository) GetListOfMerchantsWithActiveAutoWithdrawalStatus(ctx context.Context) ([]merchant.MerchantWithActiveAutoWithdrawalStatus, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetListOfMerchantsWithActiveAutoWithdrawalStatus")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "merchants, accounts, bank_accounts")

	var result []merchant.MerchantWithActiveAutoWithdrawalStatus

	rawQuery := `SELECT
		merchant_id, merchant_name, account_name,
		beneficiary->>'$.bank_code' AS beneficiary_bank_code,
		beneficiary->>'$.account_no' AS beneficiary_account_no
	FROM (
		SELECT 
			m.uuid AS merchant_id, m.name AS merchant_name, ac.name AS account_name, 
			(SELECT
				JSON_OBJECT('bank_code', beneficiary_bank_code, 'account_no', beneficiary_account_no) 
			FROM bank_accounts
			WHERE merchant_id = m.uuid AND deleted_at IS NULL LIMIT 1) AS beneficiary
		FROM merchants m
		JOIN accounts ac ON ac.reference_id = m.uuid AND ac.name IN ('PAYMENT', 'DISBURSEMENT', 'VIRTUAL_TERMINAL')
		LEFT JOIN merchant_forbidden_usecases mfu ON mfu.merchant_id = m.uuid AND mfu.use_case = 'WITHDRAWAL' AND mfu.deleted_at IS NULL
		WHERE
			m.transaction_configs->>'$.autoWithdrawal' = 'ON' AND m.status = 'ACTIVE' AND m.deleted_at IS NULL AND mfu.uuid IS NULL
	) foo 
	WHERE beneficiary IS NOT NULL ORDER BY merchant_id, account_name;`

	if err := r.db.SelectContext(ctx, &result, rawQuery); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return result, nil
}

func (r *MerchantRepository) GetListOfMerchantsToForceTheAutoWithdrawalProcess(ctx context.Context) ([]merchant.MerchantWithdrawalDetails, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/merchant/GetListOfMerchantsToForceTheAutoWithdrawalProcess")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "merchants,accounts,bank_accounts")

	rawQuery := `
		SELECT
			m.uuid AS merchant_id, m.name AS merchant_name, m.merchant_email, acc.uuid AS account_id, 
			acc.name AS account_name, ba.beneficiary_bank_code, ba.beneficiary_bank_name, ba.beneficiary_account_no, ba.beneficiary_account_name 
		FROM merchants m
		JOIN bank_accounts ba ON ba.merchant_id = m.uuid
		JOIN accounts acc ON acc.reference_id = ba.merchant_id
		LEFT JOIN merchant_forbidden_usecases mfu ON mfu.merchant_id = m.uuid AND mfu.use_case = 'WITHDRAWAL' AND mfu.deleted_at IS NULL
		WHERE 
			m.status = 'ACTIVE'
			AND m.deleted_at IS NULL
			AND m.transaction_configs->>'$.autoWithdrawal' = 'OFF'
			AND acc.name IN ('DISBURSEMENT', 'PAYMENT', 'VIRTUAL_TERMINAL')
			AND mfu.uuid IS NULL
		ORDER BY m.created_at, m.uuid;`

	result := []merchant.MerchantWithdrawalDetails{}
	if err := r.db.SelectContext(ctx, &result, rawQuery); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return result, nil
}
