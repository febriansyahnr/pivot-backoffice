package ruleevaluationsrepository

import (
	"context"
	"errors"

	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/ruleEvaluations"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *RuleEvaluationsRepository) Update(ctx context.Context, eval *ruleevaluationsmodel.RuleEvaluations) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/fraudRules/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	query := `
		UPDATE ` + tableName + `
		SET 
			reference_id = :reference_id,
			rule_id = :rule_id,
			result = :result,
			score = :score,
			reason = :reason,
			evaluated_at = :evaluated_at,
		WHERE 
			uuid = :uuid
	`
	affected, err := r.db.NamedExecContext(ctx, query, eval)
	if err != nil {
		r.logger.Error(ctx, "error when update rule evaluations", logger.Error(err), logger.Any("rule", eval))
		return err
	}

	if !affected {
		r.logger.Info(ctx, "failed when update rule evaluations", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}
	return nil
}
