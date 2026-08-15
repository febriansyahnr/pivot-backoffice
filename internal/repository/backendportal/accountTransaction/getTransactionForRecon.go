package accounttransaction_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountTransactionRepository) GetTransactionForRecon(ctx context.Context, params *reconciliation.ReconTransactionQuery) (*reconciliation.ReconTransactionModel, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetTransactionForRecon")
	defer segment.End()

	var (
		err         error
		transaction reconciliation.ReconTransactionModel
	)
	// Build channel condition based on input parameter
	var channelCondition string
	var channelArgs []interface{}

	if params.Channel == "QR" || params.Channel == "QRIS" {
		channelCondition = "and t.channel IN (?, ?)"
		channelArgs = []interface{}{"QRIS", "QR"}
	} else {
		channelCondition = "and t.channel = ?"
		channelArgs = []interface{}{params.Channel}
	}

	query := `
		select 
				t.uuid,
				t.type,
				IF(t.credit = 0, t.debit, t.credit) as amount,
				t.reference_id,
				m.name as merchant_name,
				t.processor_reference,
				t.processor_reference_id,
				t.reference,
				t.channel,
				t.status,
				t.reason_type,
				t.reason_description,
				t.additional_info,
				t.transaction_timestamp,
				COALESCE(p.type, '') as payment_type,
				COALESCE(p.processor_reference_number, '') as processor_reference_number
		from account_transactions t
		left join payments p on p.uuid = t.reference_id AND t.type = 'PAYMENT'
		left join merchants m on m.uuid = t.merchant_id
		where 
		    t.updated_at >= ? and t.updated_at < ?
		and t.reference = ?
		and t.type = ?
		and JSON_EXTRACT(t.additional_info, "$.reconReferenceNo") = ?
		` + channelCondition + `
		%s
		%s
		limit 1
	`

	settlementModelQuery := "AND (settlement_model = ? OR settlement_model IS NULL)"
	if params.SettlementModel == constant.PaymentMethodChannelTypeDirect {
		settlementModelQuery = "AND settlement_model = ?"
	}

	if params.WithTimeDuration {
		timeBefore := params.TransactionDate.Add(-params.Duration)
		timeAfter := params.TransactionDate.Add(params.Duration)
		dateTimeQuery := "and t.transaction_timestamp between ? and ?"
		query = fmt.Sprintf(query, settlementModelQuery, dateTimeQuery)

		// Build query args with channel args
		queryArgs := []interface{}{params.StartUpdatedAt, params.EndUpdatedAt, params.Reference, params.TransactionType, params.ReferenceID}
		queryArgs = append(queryArgs, channelArgs...)
		queryArgs = append(queryArgs, params.SettlementModel, timeBefore, timeAfter)

		err = r.db.GetContext(ctx, &transaction, query, queryArgs...)
	} else {
		dateTimeQuery := "and date(t.transaction_timestamp) = date(?)"
		query = fmt.Sprintf(query, settlementModelQuery, dateTimeQuery)

		// Build query args with channel args
		queryArgs := []interface{}{params.StartUpdatedAt, params.EndUpdatedAt, params.Reference, params.TransactionType, params.ReferenceID}
		queryArgs = append(queryArgs, channelArgs...)
		queryArgs = append(queryArgs, params.SettlementModel, params.TransactionDate)

		err = r.db.GetContext(ctx, &transaction, query, queryArgs...)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when find account transaction for recon", logger.Error(err))
		return nil, err
	}

	return &transaction, nil
}

func (r *AccountTransactionRepository) GetTotalPaymentAmount(ctx context.Context, params *reconciliation.PaymentTotalAmountQuery) (*reconciliation.PaymentTotalAmountResult, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/GetTransactionForRecon")
	defer segment.End()

	var (
		err          error
		totalAmount  = make(reconciliation.PaymentTotalAmountResult)
		transactions []reconciliation.ReconTransactionModel
	)

	if params.Channel != constant.ChannelVirtualAccount {
		// right now, only support for virtual account
		return &totalAmount, fmt.Errorf("channel %s is not supported", params.Channel)
	}

	startTime := params.StartTime
	endTime := params.EndTime
	duration := endTime.Sub(startTime)

	// if duration > 3 days, set start time to endtime - 3 days
	// max recon time is 3 days
	if duration > 3*24*time.Hour {
		startTime = endTime.Add(-3 * 24 * time.Hour)
	}

	// get date from params wth format YYYY-MM-DD
	startDateStr := startTime.Format("2006-01-02")
	endDateStr := endTime.Format("2006-01-02")

	query := fmt.Sprintf(`
		select 
				t.uuid,
				t.type,
				IF(t.credit = 0, t.debit, t.credit) as amount,
				t.reference_id,
				m.name as merchant_name,
				t.processor_reference,
				t.processor_reference_id,
				t.reference,
				t.channel,
				t.status,
				t.reason_type,
				t.reason_description,
				t.additional_info,
				t.transaction_timestamp,
				COALESCE(p.type, '') as payment_type,
				COALESCE(p.processor_reference_number, '') as processor_reference_number
		FROM account_transactions t 
		LEFT JOIN payments p on p.uuid = t.reference_id AND t.type = 'PAYMENT'
		LEFT JOIN merchants m ON m.uuid = t.merchant_id
		WHERE p.processor_reference_number in ('%s')
		AND date(transaction_timestamp) BETWEEN '%s' AND '%s'
		AND p.type = 'MULTIPLE'
		AND t.channel = ?
		AND t.settlement_model = 'AGGREGATOR'
		AND t.status = 'SUCCESS'
	`, params.GetReferenceIDQuery(), startDateStr, endDateStr)

	queryArgs := []interface{}{params.Channel}

	err = r.db.SelectContext(ctx, &transactions, query, queryArgs...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when find total payment amount for recon", logger.Error(err))
		return nil, err
	}

	for _, transaction := range transactions {
		totalAmount.Add(transaction.ProcessorReferenceNumber, transaction.Amount)
	}

	return &totalAmount, nil
}
