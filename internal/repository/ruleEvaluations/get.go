package ruleevaluationsrepository

import (
	"context"
	"database/sql"
	"errors"

	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/ruleEvaluations"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RuleEvaluationsRepository) GetByID(ctx context.Context, id string) (ruleEval *ruleevaluationsmodel.RuleEvaluations, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/ruleEvaluations/GetById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	ruleEval = &ruleevaluationsmodel.RuleEvaluations{}

	query := `SELECT 
		uuid, 
		reference_id,
		rule_id,
		result,
		score,
		reason,
		evaluated_at
	FROM ` + tableName + ` WHERE uuid = ?`

	err = r.db.GetContext(ctx, ruleEval, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "rule evaluations not found", logger.Any("data", map[string]string{"id": id}))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding rule evaluations", logger.Error(err))
		return ruleEval, err
	}

	return ruleEval, nil
}

func (r *RuleEvaluationsRepository) GetByRefID(ctx context.Context, refID string) (ruleEval *[]ruleevaluationsmodel.RuleEvaluations, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/ruleEvaluations/GetByRefID")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	ruleEval = &[]ruleevaluationsmodel.RuleEvaluations{}

	query := `SELECT 
		uuid, 
		reference_id,
		rule_id,
		result,
		score,
		reason,
		evaluated_at
	FROM ` + tableName + ` 
	WHERE reference_id = ?
	`

	err = r.db.SelectContext(ctx, ruleEval, query, refID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "rule evaluations not found", logger.Any("data", map[string]string{"reference_id": refID}))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding rule evaluations", logger.Error(err))
		return ruleEval, err
	}

	return ruleEval, nil
}
