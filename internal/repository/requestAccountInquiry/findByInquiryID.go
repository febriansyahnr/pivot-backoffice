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

func (r *RequestAccountInquiryRepository) FindLatestWithMasterByInquiryID(ctx context.Context, inquiryID, merchantID string) (*requestAccountInquiry.RequestAccountInquiryWithMaster, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/requestAccountInquiry/FindLatestWithMasterByInquiryID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var inquiryAccount requestAccountInquiry.RequestAccountInquiryWithMaster
	query := `
			SELECT 
				r.uuid, r.merchant_id, r.account_inquiry_id, r.beneficiary_bank_code, r.beneficiary_bank_name, r.beneficiary_account_no, 
				r.beneficiary_account_name, r.status, r.metadata, r.created_at, IFNULL(a.beneficiary_account_name, r.beneficiary_bank_name) AS master_beneficiary_account_name
			FROM ` + tableName + ` r
			LEFT JOIN account_inquiries a ON r.account_inquiry_id = a.uuid 
			WHERE r.merchant_id = ? AND r.account_inquiry_id = ? ORDER BY created_at DESC LIMIT 1`

	if err := r.db.GetContext(ctx, &inquiryAccount, query, merchantID, inquiryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "inquiry ID not found", logger.String("inquiryID", inquiryID))
			return nil, nil
		}

		r.logger.Error(ctx, "error when FindLatestWithMasterByInquiryID", logger.Error(err))
		return nil, err
	}

	if inquiryAccount.Metadata.Valid {
		_ = json.Unmarshal(inquiryAccount.Metadata.JSONText, &inquiryAccount.MetadataObj)
	}

	return &inquiryAccount, nil
}

func (r *RequestAccountInquiryRepository) FindLatestByInquiryID(ctx context.Context, inquiryID, merchantID string) (*requestAccountInquiry.RequestAccountInquiryWithMaster, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/requestAccountInquiry/FindLatestByInquiryID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var inquiryAccount requestAccountInquiry.RequestAccountInquiryWithMaster
	query := `
			SELECT 
				r.uuid, r.merchant_id, r.account_inquiry_id, r.beneficiary_bank_code, r.beneficiary_bank_name, r.beneficiary_account_no, 
				r.beneficiary_account_name, r.status, r.metadata, r.created_at
			FROM ` + tableName + ` r
			WHERE r.merchant_id = ? AND r.account_inquiry_id = ? ORDER BY created_at DESC LIMIT 1`

	if err := r.db.GetContext(ctx, &inquiryAccount, query, merchantID, inquiryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "inquiry ID not found", logger.String("inquiryID", inquiryID))
			return nil, nil
		}

		r.logger.Error(ctx, "error when FindLatestByInquiryID", logger.Error(err))
		return nil, err
	}

	if inquiryAccount.Metadata.Valid {
		_ = json.Unmarshal(inquiryAccount.Metadata.JSONText, &inquiryAccount.MetadataObj)
	}

	return &inquiryAccount, nil
}
