package accounttransaction_repository

import (
	"context"
	"errors"
	"strings"

	"github.com/paper-indonesia/pdk/v2/logger"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *AccountTransactionRepository) BulkInsert(ctx context.Context, transactions []*orchestrator_model.AccountTransaction) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/BulkInsert")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	noOfCol := 20
	values := make([]string, 0, len(transactions))
	args := make([]interface{}, 0, len(transactions)*noOfCol)
	for _, val := range transactions {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args, val.UUID, val.ReferenceID, val.MerchantID, val.AccountID, val.Currency, val.Credit, val.Debit,
			val.Type, val.Reference, val.Channel, val.Status, val.ReasonType, val.ReasonDescription, val.Remarks, val.TransactionTimestamp,
			val.AdditionalInfo, val.SettlementAt, val.SettlementStatus, val.Processor, val.ProcessorID, val.ProcessorTransactionID, val.CreatedAt, val.UpdatedAt, val.MerchantReferenceID)
	}

	query := `
		INSERT INTO ` + tableName + `(
			uuid, reference_id, merchant_id, account_id, 
			currency, credit, debit, type, reference, channel,
			status, reason_type, reason_description, remarks, 
			transaction_timestamp, additional_info, settlement_at, settlement_status, processor_reference,
			processor_reference_id, processor_transaction_id, created_at, updated_at, merchant_reference_id
		) VALUES ` + strings.Join(values, ",")
	affected, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error(ctx, "error when bulk insert into account_transactions", logger.Error(err), logger.Any("query", query), logger.Any("args", args))
		return err
	}

	if !affected {
		err := errors.New("no rows affected")
		r.logger.Error(ctx, "failed when bulk insert into account_transactions", logger.Error(err))
		return err
	}

	return nil
}
