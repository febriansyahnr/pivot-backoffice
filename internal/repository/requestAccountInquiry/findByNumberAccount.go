package requestaccountinquiry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	requestAccountInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RequestAccountInquiryRepository) FindLatestByNumberAccount(ctx context.Context, accountNo, merchantID string) (*requestAccountInquiry.RequestAccountInquiries, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/v1/requestAccountInquiry/FindLatestByNumberAccount")
	defer span.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := `SELECT 
				r.uuid, r.merchant_id, r.account_inquiry_id, r.beneficiary_bank_code, r.beneficiary_bank_name, r.beneficiary_account_no, 
				r.beneficiary_account_name, r.status, r.metadata, r.created_at
			FROM ` + tableName + ` r
			WHERE r.merchant_id = ? AND r.beneficiary_account_no = ? ORDER BY created_at DESC LIMIT 1`

	var inquiryAccount requestAccountInquiry.RequestAccountInquiries
	err := r.db.GetContext(ctx, &inquiryAccount, query, merchantID, accountNo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when FindLatestByNumberAccount", logger.Error(err))
		return nil, err
	}

	if inquiryAccount.Metadata.Valid {
		_ = json.Unmarshal(inquiryAccount.Metadata.JSONText, &inquiryAccount.MetadataObj)
	}

	return &inquiryAccount, nil
}
