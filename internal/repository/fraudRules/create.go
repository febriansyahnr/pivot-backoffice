package fraudrulesrepository

import (
	"context"

	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *FraudRulesRepository) Create(ctx context.Context, rules *fraudrulesmodel.FraudRules) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/fraudRules/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)
	query := "INSERT INTO " + tableName + " " +
		"(uuid, rule_name, `condition`, priority, weight, is_active, reference_type, provider, created_at, updated_at, deleted_at) " +
		"VALUES (:uuid, :rule_name, :condition, :priority, :weight, :is_active, :reference_type, :provider, :created_at, :updated_at, :deleted_at)"

	_, err := r.db.NamedExecContext(ctx, query, rules)
	if err != nil {
		r.logger.Error(ctx, "error when inserting fraud rules", logger.Error(err), logger.Any("configuration", rules))
		return err
	}
	return nil
}
