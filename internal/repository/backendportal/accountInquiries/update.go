package accountInquiries

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/accountInquiries"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountInquiriesRepository) Update(ctx context.Context, account *accountInquiries.AccountInquiries) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/AccountInquiries/Update")
	defer segment.End()

	query := `
		UPDATE account_inquiries
		SET
			beneficiary_bank_code = :beneficiary_bank_code,
			beneficiary_bank_name = :beneficiary_bank_name,
			beneficiary_account_no = :beneficiary_account_no,
			beneficiary_account_name = :beneficiary_account_name,
			response = :response,
			updated_at = :updated_at
		WHERE
			uuid = :uuid
	`

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	_, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		r.logger.Error(ctx, "error when updating account inquiries", logger.Error(err))
		return err
	}
	return nil
}
