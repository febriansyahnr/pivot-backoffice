package accounttransaction_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountTransactionRepository) UpdateStatusAccountTransaction(ctx context.Context, id string, status string, reasonType, reasonDescription *string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateStatusAccountTransaction")
	defer segment.End()
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE account_transactions
		SET
			status = ?, 
			reason_type = ?,
			reason_description = ?,
			updated_at = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		status,
		reasonType,
		reasonDescription,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating status account_transactions", logger.Error(err))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) UpdateReasonOnly(ctx context.Context, id string, reasonType, reasonDescription *string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateReasonOnly")
	defer segment.End()
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE account_transactions
		SET
			reason_type = ?,
			reason_description = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		reasonType,
		reasonDescription,
		id,
	)
	if err != nil {
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating reason only account_transactions", logger.Error(err))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) UpdateStatusAccountTransactionByReferenceID(ctx context.Context, id string, status string, reasonType, reasonDescription *string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateStatusAccountTransactionByReferenceID")
	defer segment.End()
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE account_transactions
		SET
			status = ?, 
			reason_type = ?,
			reason_description = ?,
			updated_at = ?
		WHERE reference_id = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		status,
		reasonType,
		reasonDescription,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating status account_transactions", logger.Error(err))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) UpdateProcessorAndReconReference(
	ctx context.Context,
	id string,
	processorReferenceName, processorReferenceId, reconReferenceNo string,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateProcessorAndReconReference")
	defer segment.End()

	fields := []string{"processor_reference = ?", "processor_reference_id = ?", "updated_at = ?"}
	values := []interface{}{processorReferenceName, processorReferenceId, time.Now().UTC()}
	jsonFieldSet := []string{}

	if reconReferenceNo != "" {
		jsonFieldSet = append(jsonFieldSet, "'$.reconReferenceNo', ?")
		values = append(values, reconReferenceNo)
	}

	// If any JSON fields exist, set them all in one JSON_SET clause
	if len(jsonFieldSet) > 0 {
		fields = append(fields, fmt.Sprintf("additional_info = JSON_SET(COALESCE(additional_info, '{}'), %s)", strings.Join(jsonFieldSet, ", ")))
	}

	// Final query string
	rawQuery := `
		UPDATE account_transactions
		SET ` + strings.Join(fields, ", ") + `
		WHERE uuid = ?;`
	values = append(values, id)

	_, err := r.db.ExecContext(
		ctx,
		rawQuery,
		values...,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating status account_transactions", logger.Error(err))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) CancelIndirectTransactionFee(ctx context.Context, id string, date time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/CancelIndirectTransactionFee")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, "account_transactions")

	rawQuery := `UPDATE account_transactions 
		SET additional_info = JSON_SET(additional_info, '$.canceledAt', ?) WHERE uuid = ?;`

	_, err := r.db.ExecContext(ctx, rawQuery, date, id)
	return err
}

func (r *AccountTransactionRepository) DeductBalanceForIndirectFeeType(ctx context.Context, merchantId string, ids []string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/DeductBalanceForIndirectFeeType")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `UPDATE account_transactions 
		SET
			status = 'SUCCESS', updated_at = NOW()
		WHERE merchant_id = ? AND uuid IN (?) AND status = 'PENDING';`

	query, args, err := r.db.In(rawQuery, merchantId, ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *AccountTransactionRepository) UpdateAdditionalInfoByID(ctx context.Context, id string, additionalInfo types.NullJSONText) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateStatusAccountTransaction")
	defer segment.End()
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE account_transactions
		SET
			additional_info = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		additionalInfo,
		id,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating status account_transactions", logger.Error(err))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) UpdateSettlementStatusAndSettlementAtByID(ctx context.Context, id string, settlementStatus string, settlementAt time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateSettlementStatusAndSettlementAtByID")
	defer segment.End()
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE account_transactions
		SET
			settlement_status = ?,
			settlement_at = ?,
			updated_at = ?
		WHERE uuid = ?
		`
	_, err := r.db.ExecContext(
		ctx,
		query,
		settlementStatus,
		settlementAt,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating status account_transactions", logger.Error(err))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) UpdateTransactionsStatusAndAdditionalInfoByID(ctx context.Context, id string, status string, reasonType string, reasonDescription string, additionalInfo types.NullJSONText) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateAdditionalInfoAndStatusByID")
	defer segment.End()
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE account_transactions
		SET
			status = ?,
			reason_type = ?,
			reason_description = ?,
			additional_info = ?,
			updated_at = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		status,
		reasonType,
		reasonDescription,
		additionalInfo,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating status account_transactions", logger.Error(err))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) VoidTransaction(ctx context.Context, request *orchestrator_model.VoidTransactionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/VoidTransaction")
	defer segment.End()
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE account_transactions
		SET
			status = ?,
			reason_type = ?,
			reason_description = ?,
			settlement_status = ?,
			updated_at = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		request.Status,
		request.ReasonType,
		request.ReasonDescription,
		request.SettlementStatus,
		time.Now().UTC(),
		request.TrxID,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when void account_transactions", logger.Error(err), logger.Any("request", request))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) UpdateTransactionTimestamp(ctx context.Context, id string, transactionTimestamp time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateTransactionTimestamp")
	defer segment.End()
	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := `
		UPDATE account_transactions
		SET
			transaction_timestamp = ?,
			updated_at = ?
		WHERE uuid = ?
		`

	_, err := r.db.ExecContext(
		ctx,
		query,
		transactionTimestamp,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating transaction timestamp account_transactions", logger.Error(err))
		return err
	}

	return nil
}

func (r *AccountTransactionRepository) RearrangeUpdatedAtForTransactionWithPendingStatus(ctx context.Context, referenceIds []string, updatedAt time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/RearrangeUpdatedAtForTransactionWithPendingStatus")
	defer segment.End()

	if len(referenceIds) == 0 {
		return nil
	}

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `UPDATE
		account_transactions
	SET
		updated_at = ?
	WHERE reference_id IN (?) AND status = 'PENDING';`

	query, args, err := r.db.In(rawQuery, updatedAt, referenceIds)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *AccountTransactionRepository) UpdateTransactionWithPendingStatusByReferenceIdTypeAndChannel(ctx context.Context, referenceId, typ, currentChannel string, data orchestrator_model.UpdateTransactionWithPendingStatus) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/UpdateTransactionWithPendingStatusByReferenceIdTypeAndChannel")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `UPDATE
		account_transactions
	SET 
		additional_info = ?, processor_reference_id = ?, processor_reference = ?, updated_at = ?, channel = ?
	WHERE 
		reference_id = ? AND type = ? AND channel = ? AND status = 'PENDING';`

	args := []interface{}{
		types.JSONText(data.Metadata), data.ProcessorID, data.Processor, data.UpdatedAt, data.Channel, referenceId, typ, currentChannel,
	}

	if affected, err := r.db.ExecContext(ctx, rawQuery, args...); err != nil {
		return err

	} else if !affected {
		return fmt.Errorf("update metadata for transaction with pending status: %w", constant.ErrDataNotFound)
	}
	return nil
}

func (r *AccountTransactionRepository) UpdatePaymentTransactionStatusAndMetadataByID(ctx context.Context, request orchestrator_model.UpdatePaymentTransactionRequest, metadata orchestrator_model.MetadataPayment[any]) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/UpdatePaymentTransactionStatusAndMetadataByID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	fields, values, jsonFieldSet := []string{}, []interface{}{}, []string{}

	if request.Status != "" {
		fields = append(fields, "status = ?")
		values = append(values, request.Status)
	}
	if !request.UpdatedAt.IsZero() {
		fields = append(fields, "updated_at = ?")
		values = append(values, request.UpdatedAt)
	} else {
		fields = append(fields, "updated_at = ?")
		values = append(values, time.Now().UTC())
	}
	if !request.TransactionTimestamp.IsZero() {
		fields = append(fields, "transaction_timestamp = ?")
		values = append(values, request.TransactionTimestamp)
	}
	if request.ProcessorReferenceName != "" {
		fields = append(fields, "processor_reference = ?")
		values = append(values, request.ProcessorReferenceName)
	}
	if request.ProcessorReferenceId != "" {
		fields = append(fields, "processor_reference_id = ?")
		values = append(values, request.ProcessorReferenceId)
	}
	if request.ProcessorTransactionId != "" {
		fields = append(fields, "processor_transaction_id = ?")
		values = append(values, request.ProcessorTransactionId)
	}
	if request.SettlementStatus != nil {
		fields = append(fields, "settlement_status = ?")
		values = append(values, *request.SettlementStatus)
	}
	if request.SettlementAt != nil && !request.SettlementAt.IsZero() {
		fields = append(fields, "settlement_at = ?")
		values = append(values, *request.SettlementAt)
	}
	if request.SettlementModel != nil {
		fields = append(fields, "settlement_model = ?")
		values = append(values, *request.SettlementModel)
	}
	if metadata.ReconReferenceNo != "" {
		jsonFieldSet = append(jsonFieldSet, "'$.reconReferenceNo', ?")
		values = append(values, metadata.ReconReferenceNo)
	}
	if metadata.FailureCode != "" {
		jsonFieldSet = append(jsonFieldSet, "'$.failureCode', ?")
		values = append(values, metadata.FailureCode)
	}
	if metadata.SettlementDetail != nil {
		buf, _ := json.Marshal(metadata.SettlementDetail)
		jsonFieldSet = append(jsonFieldSet, "'$.settlementDetail', CAST(? AS JSON)")
		values = append(values, buf)
	}
	if metadata.MethodDetail != nil {
		buf, _ := json.Marshal(metadata.MethodDetail)
		jsonFieldSet = append(jsonFieldSet, "'$.methodDetail', CAST(? AS JSON)")
		values = append(values, buf)
	}
	if metadata.ChargeStatus != "" {
		buf, _ := json.Marshal(metadata.ChargeStatus)
		jsonFieldSet = append(jsonFieldSet, "'$.chargeStatus', CAST(? AS JSON)")
		values = append(values, buf)
	}
	if metadata.FeeDetail != nil {
		buf, _ := json.Marshal(metadata.FeeDetail)
		jsonFieldSet = append(jsonFieldSet, "'$.feeDetail', CAST(? AS JSON)")
		values = append(values, buf)
	}
	if metadata.FeeOnBehalf != nil {
		buf, _ := json.Marshal(metadata.FeeOnBehalf)
		jsonFieldSet = append(jsonFieldSet, "'$.feeOnBehalf', CAST(? AS JSON)")
		values = append(values, buf)
	}
	if metadata.ReconDetail != nil {
		buf, _ := json.Marshal(metadata.ReconDetail)
		jsonFieldSet = append(jsonFieldSet, "'$.reconDetail', CAST(? AS JSON)")
		values = append(values, buf)
	}
	if metadata.StatementDescriptor != "" {
		buf, _ := json.Marshal(metadata.StatementDescriptor)
		jsonFieldSet = append(jsonFieldSet, "'$.statementDescriptor', CAST(? AS JSON)")
		values = append(values, buf)
	}
	if !metadata.ExpiredAt.IsZero() {
		buf, _ := json.Marshal(metadata.ExpiredAt)
		jsonFieldSet = append(jsonFieldSet, "'$.expiredAt', CAST(? AS JSON)")
		values = append(values, buf)
	}

	if metadata.SubPaymentSummary != nil {
		buf, _ := json.Marshal(metadata.SubPaymentSummary)
		jsonFieldSet = append(jsonFieldSet, "'$.subPaymentSummary', CAST(? AS JSON)")
		values = append(values, buf)
	}

	if len(jsonFieldSet) > 0 {
		fields = append(fields, fmt.Sprintf("additional_info = JSON_SET(additional_info, %s)", strings.Join(jsonFieldSet, ", ")))
	}

	rawQuery := `UPDATE
		account_transactions
	SET 
		` + strings.Join(fields, ", ") + ` 
	WHERE
		uuid = ?;`
	values = append(values, request.LedgerId)

	if affected, err := r.db.ExecContext(ctx, rawQuery, values...); err != nil {
		return err

	} else if !affected {
		return constant.ErrDataNotFound
	}
	return nil
}

func (r *AccountTransactionRepository) BulkUpdateTransactions(ctx context.Context, request []*orchestrator_model.AccountTransaction) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/BulkUpdateTransactions")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	ctx, err := r.db.BeginTxx(ctx)
	if err != nil {
		r.logger.Error(ctx, "error when starting bulk update transactions", logger.Error(err))
		return err
	}

	for _, trx := range request {
		query := `UPDATE account_transactions
					SET
						status = ?,
						reason_type = ?,
						reason_description = ?,
						settlement_status = ?,
						updated_at = ?,
						reference = ?,
						processor_reference = ?,
						processor_reference_id = ?,
						processor_transaction_id = ?
					WHERE uuid = ?`

		_, err := r.db.ExecContext(ctx, query, trx.Status, trx.ReasonType.String, trx.ReasonDescription.String, trx.SettlementStatus.String, trx.UpdatedAt, trx.Reference, trx.Processor, trx.ProcessorID, trx.ProcessorTransactionID, trx.UUID)
		if err != nil {
			r.logger.Error(ctx, "error when bulk update transactions", logger.Error(err), logger.Any("transaction", trx))
			errTrx := r.db.Rollback(ctx)
			if errTrx != nil {
				r.logger.Error(ctx, "error when rolling back bulk update transactions", logger.Error(errTrx))
			}
			return err
		}
	}

	err = r.db.Commit(ctx)
	if err != nil {
		r.logger.Error(ctx, "error when commit bulk update transactions", logger.Error(err))
		return err
	}
	return nil
}

func (r *AccountTransactionRepository) UpdateFDSRiskAssessmentResultByID(ctx context.Context, id string, data fdscommon.RiskAssessmentResult) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateFDSRiskAssessmentResultByID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	rawQuery := `UPDATE
		account_transactions
	SET 
		additional_info = JSON_SET(IFNULL(additional_info, JSON_OBJECT()), '$.fdsRiskAssessment', CAST(? AS JSON))
	WHERE
		uuid = ?;`
	if _, err := r.db.ExecContext(ctx, rawQuery, data.RawResult(), id); err != nil {
		return err
	}
	return nil
}

func (r *AccountTransactionRepository) UpdateSettlementDetailByIDs(ctx context.Context, ids []string, request orchestrator_model.UpdateSettlementDetailRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateSettlementDetailByIDs")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	params := make([]string, 0, 1)
	values := make([]any, 0, 1)

	if request.EstimateSettlementAt != nil {
		values = append(values, request.EstimateSettlementAt.Format(time.RFC3339))
		params = append(params, "additional_info = JSON_SET(additional_info, '$.settlementDetail.estimateSettlementAt', ?)")
	}

	if len(params) == 0 {
		return errors.New("at least one settlement detail field must be provided for update")
	}
	args := append(values, ids)

	rawQuery := fmt.Sprintf(`UPDATE account_transactions SET %s WHERE uuid IN (?);`, strings.Join(params, ", "))

	rawQuery, args, _ = sqlx.In(rawQuery, args...)
	if _, err := r.db.ExecContext(ctx, rawQuery, args...); err != nil {
		return err
	}
	return nil
}

func (r *AccountTransactionRepository) UpdateSettlementHoldByReferenceID(ctx context.Context, referenceId string, flag bool) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateSettlementHoldByReferenceID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)
	flagStr := "false"
	if flag {
		flagStr = "true"
	}
	rawQuery := `UPDATE
		account_transactions
	SET 
		additional_info = JSON_SET(IFNULL(additional_info, JSON_OBJECT()), '$.settlementDetail.isOnHold', CAST(? AS JSON))
	WHERE
		reference_id = ? AND status = 'SUCCESS' AND settlement_status = 'PENDING';`
	if _, err := r.db.ExecContext(ctx, rawQuery, flagStr, referenceId); err != nil {
		r.logger.Error(ctx, "error update settlement hold by reference id", logger.Error(err))
		return err
	}
	return nil
}

func (r *AccountTransactionRepository) UpdateTransactionDetail(ctx context.Context, request orchestrator_model.UpdateTransactionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateTransactionDetail")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)
	updateSetClause := []string{}
	updateSetClauseValue := []any{}

	if request.Channel != "" {
		updateSetClause = append(updateSetClause, "channel = ?")
		updateSetClauseValue = append(updateSetClauseValue, request.Channel)
	}
	if len(updateSetClause) == 0 {
		return errors.New("no fields to update")
	}
	rawQuery := fmt.Sprintf(`UPDATE account_transactions SET %s WHERE uuid = ?;`, strings.Join(updateSetClause, ", "))
	updateSetClauseValue = append(updateSetClauseValue, request.TransactionID)

	if _, err := r.db.ExecContext(ctx, rawQuery, updateSetClauseValue...); err != nil {
		r.logger.Error(ctx, "error update transaction detail", logger.Error(err))
		return err
	}
	return nil
}
