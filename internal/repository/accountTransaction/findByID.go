package accounttransaction_repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *AccountTransactionRepository) FindByID(ctx context.Context, id string) (*orchestratorModel.AccountTransactionWithUseCase, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/FindByID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	var transaction orchestratorModel.AccountTransactionWithUseCase
	query := queryWithUseCase + ` WHERE t.uuid = ?`

	if err := r.db.GetContext(ctx, &transaction, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		r.logger.Error(ctx, "error when find account transaction by id", logger.Error(err))
		return nil, err
	}

	if transaction.AdditionalInfo.Valid {
		detail := orchestratorModel.FeeTransactionMetadataObject{}
		_ = json.Unmarshal(transaction.AdditionalInfo.JSONText, &detail)
		transaction.AdditionalInfoObj = detail
	}
	return &transaction, nil
}
