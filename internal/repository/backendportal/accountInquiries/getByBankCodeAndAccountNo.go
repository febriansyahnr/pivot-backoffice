package accountInquiries

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/accountInquiries"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *AccountInquiriesRepository) GetByBankCodeAndAccountNo(
	ctx context.Context, bankCode, accountNo string) (*accountInquiries.AccountInquiries, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/AccountInquiries/GetByBankCodeAndAccountNo")
	defer segment.End()

	var account accountInquiries.AccountInquiries

	// QUERY CONDITION
	query := `SELECT 
    	uuid, 
    	beneficiary_bank_code, 
    	beneficiary_bank_name, 
    	beneficiary_account_no, 
    	beneficiary_account_name, 
    	response,
    	created_at, 
    	updated_at 
	FROM account_inquiries
	WHERE beneficiary_bank_code = ? AND beneficiary_account_no = ?`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	if err := r.db.GetContext(ctx, &account, query, bankCode, accountNo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "account inquiry not found", logger.Any("data", map[string]string{"bank_code": bankCode, "account_no": accountNo}))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding account inquiry", logger.Error(err))
		return &account, err
	}

	return &account, nil
}
