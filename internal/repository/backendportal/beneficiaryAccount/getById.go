package beneficiaryAccountRepository

import (
	"context"
	"database/sql"
	"errors"

	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *BeneficiaryAccountRepository) GetByID(ctx context.Context, id string) (*beneficiaryAccountModel.BeneficiaryAccount, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/beneficiaryAccount/GetById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "beneficiary_accounts")

	var account beneficiaryAccountModel.BeneficiaryAccount

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
	WHERE uuid = ?`

	err := r.db.GetContext(ctx, &account, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "Beneficiary Account not found", logger.Any("data", map[string]string{"id": id}))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding beneficiary account", logger.Error(err))
		return &account, err
	}

	return &account, nil
}
