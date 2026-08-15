package accounttransaction_repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ledger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
)

func (r *AccountTransactionRepository) UpdateTransactionsStatus(ctx context.Context, request *ledger_model.UpdateLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/accountTransaction/UpdateStatusAccountTransaction")
	defer segment.End()
	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	setClauses := []string{"updated_at = ?"}
	args := []interface{}{time.Now().UTC()}

	if request.ReasonType != "" {
		setClauses = append(setClauses, "reason_type = ?")
		if strings.ToLower(request.ReasonType) == "null" {
			args = append(args, sql.NullString{})

		} else {
			args = append(args, request.ReasonType)
		}
	}

	if request.ReasonDescription != "" {
		setClauses = append(setClauses, "reason_description = ?")
		if strings.ToLower(request.ReasonDescription) == "null" {
			args = append(args, sql.NullString{})

		} else {
			args = append(args, request.ReasonDescription)
		}
	}

	if request.AdditionalInfo != nil {
		setClauses = append(setClauses, "additional_info = ?")
		jsonAdditionalInfo, err := json.Marshal(request.AdditionalInfo)
		if err != nil {
			r.logger.Error(ctx, "error when marshalling additional info", logger.Error(err))
			return err
		}
		args = append(args, types.NullJSONText{
			Valid:    true,
			JSONText: jsonAdditionalInfo,
		})
	}

	if request.Status != "" {
		setClauses = append(setClauses, "status = ?")
		args = append(args, request.Status)
	}

	if request.ProcessorReference != "" {
		setClauses = append(setClauses, "processor_reference = ?")
		args = append(args, request.ProcessorReference)
	}

	if request.ProcessorReferenceID != "" {
		setClauses = append(setClauses, "processor_reference_id = ?")
		args = append(args, request.ProcessorReferenceID)
	}

	if request.ProcessorTransactionID != "" {
		setClauses = append(setClauses, "processor_transaction_id = ?")
		args = append(args, request.ProcessorTransactionID)
	}

	if request.SettlementStatus != "" {
		setClauses = append(setClauses, "settlement_status = ?")
		args = append(args, request.SettlementStatus)
	}

	if !request.SettlementAt.IsZero() {
		setClauses = append(setClauses, "settlement_at = ?")
		args = append(args, request.SettlementAt)
	}

	args = append(args, request.ReferenceID)

	query := `
		UPDATE account_transactions
		SET ` + strings.Join(setClauses, ", ") + `
		WHERE reference_id = ? `

	if request.Conditional != nil {
		if request.Conditional.CurrentStatus != "" {
			query += `AND status = '` + request.Conditional.CurrentStatus + `'`
		} else {
			query += `AND status = '` + constant.StatusPending + `'`
		}

		if request.Conditional.Type != "" {
			query += ` AND type = '` + request.Conditional.Type + `'`
		}
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		// if no rows affected, return nil
		if errors.Is(err, constant.ErrNoRowsAffected) {
			return nil
		}

		r.logger.Error(ctx, "error when updating status account_transactions",
			logger.Error(err),
			logger.Any("request", request),
			logger.Any("query", query))
		return err
	}

	return nil

}
