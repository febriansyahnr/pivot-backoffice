package accountInquiries

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/accountInquiries"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountInquiriesRepository) Create(ctx context.Context, account *accountInquiries.AccountInquiries) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/AccountInquiries/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		INSERT INTO account_inquiries (uuid, beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no, beneficiary_account_name, response, created_at, updated_at, deleted_at)
		VALUES (:uuid, :beneficiary_bank_code, :beneficiary_bank_name, :beneficiary_account_no, :beneficiary_account_name, :response, :created_at, :updated_at, :deleted_at)
	`

	affected, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		r.logger.Error(ctx, "error when inserting account inquiries", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Error(ctx, "failed when inserting account inquiries", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}

	return nil
}
