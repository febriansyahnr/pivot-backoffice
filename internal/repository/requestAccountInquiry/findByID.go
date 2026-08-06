package requestaccountinquiry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	requestAccountInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RequestAccountInquiryRepository) FindByID(ctx context.Context, id string) (*requestAccountInquiry.RequestAccountInquiryWithMaster, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/requestAccountInquiry/FindLatestByInquiryID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConstant.CtxSQLTableNameKey, tableName)

	var inquiryAccount requestAccountInquiry.RequestAccountInquiryWithMaster
	query := `
			SELECT 
				r.uuid, r.merchant_id, r.account_inquiry_id, r.beneficiary_bank_code, r.beneficiary_bank_name, r.beneficiary_account_no, 
				r.beneficiary_account_name, r.status, r.metadata, r.created_at
			FROM ` + tableName + ` r
			WHERE r.uuid = ? ORDER BY created_at DESC LIMIT 1`

	if err := r.db.GetContext(ctx, &inquiryAccount, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "inquiry ID not found", logger.String("inquiryID", id))
			return nil, nil
		}

		r.logger.Error(ctx, "error when FindByID", logger.Error(err))
		return nil, err
	}

	if inquiryAccount.Metadata.Valid {
		_ = json.Unmarshal(inquiryAccount.Metadata.JSONText, &inquiryAccount.MetadataObj)
	}

	return &inquiryAccount, nil
}
