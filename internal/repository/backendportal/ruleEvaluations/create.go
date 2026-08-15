package ruleevaluationsrepository

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/logger"
	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ruleEvaluations"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *RuleEvaluationsRepository) Create(ctx context.Context, eval *ruleevaluationsmodel.RuleEvaluations) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/ruleEvaluations/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	query := `
		INSERT INTO ` + tableName + `
		(uuid, reference_id, rule_id, result, score, reason, evaluated_at) 
		VALUES (:uuid, :reference_id, :rule_id, :result, :score, :reason, :evaluated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, eval)
	if err != nil {
		r.logger.Error(ctx, "error when inserting rule evaluations", logger.Error(err), logger.Any("configuration", eval))
		return err
	}
	return nil
}
