package beneficiaryAccountRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *BeneficiaryAccountRepository) GetByBankCodeAndAccountNo(
	ctx context.Context,
	merchantId, bankCode, accountNo string) (*beneficiaryAccountModel.BeneficiaryAccount, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/beneficiaryAccount/GetByBankCodeAndAccountNo")
	defer segment.End()

	var account beneficiaryAccountModel.BeneficiaryAccount

	// QUERY CONDITION
	query := `SELECT 
    	uuid, 
    	merchant_id,
    	beneficiary_bank_code, 
    	beneficiary_bank_name, 
    	beneficiary_account_no, 
    	beneficiary_account_name, 
    	metadata,
    	created_at, 
    	updated_at 
	FROM beneficiary_accounts
	WHERE beneficiary_bank_code = ? AND beneficiary_account_no = ? AND merchant_id = ?`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "beneficiary_accounts")

	if err := r.db.GetContext(ctx, &account, query, bankCode, accountNo, merchantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "Beneficiary Account not found", logger.Any("data", map[string]string{"bank_code": bankCode, "account_no": accountNo}))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding Beneficiary Account", logger.Error(err))
		return &account, err
	}

	if account.Metadata.Valid {
		_ = json.Unmarshal(account.Metadata.JSONText, &account.MetadataObj)
	}

	return &account, nil
}
