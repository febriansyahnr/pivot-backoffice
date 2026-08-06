package requestaccountinquiry

import (
	"context"
	"errors"

	requestAccountInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RequestAccountInquiryRepository) Create(ctx context.Context, data *requestAccountInquiry.RequestAccountInquiries) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/requestAccountInquiry/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
	INSERT INTO request_account_inquiries(uuid, merchant_id, account_inquiry_id, beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no, beneficiary_account_name, status, metadata, created_at)
	VALUES(:uuid, :merchant_id, :account_inquiry_id, :beneficiary_bank_code, :beneficiary_bank_name, :beneficiary_account_no, :beneficiary_account_name, :status, :metadata, :created_at)
	`
	affected, err := r.db.NamedExecContext(ctx, query, data)
	if err != nil {
		r.logger.Error(ctx, "error when inserting request account inquiries", logger.Error(err))
		return err
	}

	if !affected {
		err = errors.New("no rows affected")
		r.logger.Error(ctx, "failed when inserting request account inquiries", logger.Error(err))
		return err
	}

	return nil
}
