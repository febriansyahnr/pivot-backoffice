package vccSettlement

import (
	"context"
	"errors"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *VccSettlementRepository) BulkInsert(ctx context.Context, data []*vccSettlement.VccSettlement) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/vccSettlement/BulkInsert")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	var (
		totalCol         = 18
		args             = make([]interface{}, 0, totalCol*len(data))
		argParamTemplate = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		argsPlaceHolder  = make([]string, len(data))
	)

	for i, val := range data {
		argsPlaceHolder[i] = argParamTemplate
		args = append(args, val.UUID, val.RcnId, val.AcquirerReferenceNumber, val.Status, val.ReferenceNo,
			val.AuthorizationNo, val.PostingDate, val.BillingCycle, val.SourceAmount, val.BillingAmount, val.TransactionDate,
			val.SettlementDate, val.MerchantName, val.MerchantCountry, val.MerchantCategory, val.CreatedAt, val.UpdatedAt, val.DeletedAt)
	}
	query := `
		INSERT INTO ` + tableName + `(
			uuid,
			rcn_id,
			acquirer_reference_number,
			status,
			reference_no,
			authorization_no,
			posting_date,
			billing_cycle,
			source_amount,
			billing_amount,
			transaction_date,
			settlement_date,
			merchant_name,
			merchant_country,
			merchant_category,
			created_at,
			updated_at,
			deleted_at
		) VALUES ` + strings.Join(argsPlaceHolder, ",")
	isInserted, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error(ctx, "error when bulk insert into vcc settlements table", logger.Error(err))
		return err
	}
	if !isInserted {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when bulk insert into vcc settlements table", logger.Error(err))
		return err
	}
	return nil

}
