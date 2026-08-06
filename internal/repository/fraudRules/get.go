package fraudrulesrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const listAllQuery = `
	SELECT %s
	FROM %s fr
	WHERE fr.is_active = 1 %s
	AND fr.deleted_at IS NULL
	ORDER BY fr.priority ASC
	%s
	`

const getByIdQuery = `
	SELECT %s
	FROM %s fr
	WHERE fr.uuid = ?
	AND fr.deleted_at IS NULL
	`

func (r *FraudRulesRepository) List(ctx context.Context, q *fraudrulesmodel.FraudRulesQuery) ([]*fraudrulesmodel.FraudRules, int, error) {
	ctx, span := otelTracer.Start(ctx, "repository/transferConfig/ListRouting")
	defer span.End()

	ctx = context.WithValue(ctx, constant.CtxSQLTableNameKey, tableName)

	// construct where condition
	whereStr := ""
	qStr := q.String()
	if qStr != "" {
		whereStr = fmt.Sprintf("AND %s", qStr)
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s fr
		WHERE fr.is_active = 1 %s
		AND fr.deleted_at IS NULL
	`, tableName, whereStr)

	var total int
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.logger.Error(ctx, "error when getting count of fraud rules", logger.Error(err))
		return nil, 0, err
	}

	limitOffsetStr := ""
	if q.PageSize > 0 {
		offset := 0
		if q.Page > 1 {
			offset = (int(q.Page) - 1) * int(q.PageSize)
		}
		limitOffsetStr = fmt.Sprintf("LIMIT %d OFFSET %d", q.PageSize, offset)
	}

	query := fmt.Sprintf(listAllQuery, tableColumns, tableName, whereStr, limitOffsetStr)
	rules := []*fraudrulesmodel.FraudRules{}

	err = r.db.SelectContext(ctx, &rules, query)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		r.logger.Error(ctx, "error when getting list of fraud rules", logger.Error(err))
		return rules, 0, err
	}
	return rules, total, nil
}

func (r *FraudRulesRepository) GetByID(ctx context.Context, id string) (rule *fraudrulesmodel.FraudRules, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/fraudRules/GetById")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, tableName)

	query := fmt.Sprintf(getByIdQuery, tableColumns, tableName)
	rule = &fraudrulesmodel.FraudRules{}

	err = r.db.GetContext(ctx, rule, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Info(ctx, "fraud rule not found", logger.Any("data", map[string]string{"id": id}))
			return nil, nil
		}

		r.logger.Error(ctx, "error when finding fraud rule", logger.Error(err))
		return rule, err
	}

	return rule, nil
}
