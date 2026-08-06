package fraudrulesrepository

import (
	"context"
	"errors"

	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *FraudRulesRepository) Update(ctx context.Context, rule *fraudrulesmodel.FraudRules) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/fraudRules/Update")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	query := "UPDATE " + tableName + " SET " +
		"rule_name = :rule_name, " +
		"`condition` = :condition, " +
		"priority = :priority, " +
		"weight = :weight, " +
		"is_active = :is_active, " +
		"reference_type = :reference_type, " +
		"provider = :provider, " +
		"updated_at = :updated_at " +
		"WHERE uuid = :uuid AND deleted_at IS NULL"

	affected, err := r.db.NamedExecContext(ctx, query, rule)
	if err != nil {
		r.logger.Error(ctx, "error when update fraud rules", logger.Error(err), logger.Any("rule", rule))
		return err
	}

	if !affected {
		r.logger.Info(ctx, "failed when update fraud rules", logger.Error(errors.New("no rows affected")))
		return errors.New("no rows affected")
	}
	return nil
}
