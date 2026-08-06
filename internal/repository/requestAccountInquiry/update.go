package requestaccountinquiry

import (
	"context"

	requestAccountInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RequestAccountInquiryRepository) Update(ctx context.Context, data *requestAccountInquiry.RequestAccountInquiryWithMaster) error {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/requestAccountInquiry/Update")
	defer span.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE request_account_inquiries
		SET
			beneficiary_bank_code = :beneficiary_bank_code,
			beneficiary_bank_name = :beneficiary_bank_name,
			beneficiary_account_no = :beneficiary_account_no,
			beneficiary_account_name = :beneficiary_account_name,
			status = :status,
			metadata = :metadata
		WHERE
			uuid = :uuid
	`

	_, err := r.db.NamedExecContext(ctx, query, data)
	if err != nil {
		r.logger.Error(ctx, "error when updating request account inquiries", logger.Error(err))
		return err
	}

	return nil
}
