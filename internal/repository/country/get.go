package countryRepository

import (
	"context"
	"database/sql"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	countryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/country"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *countryRepository) GetAll(ctx context.Context, filter *countryModel.SearchFilterRequest) ([]*countryModel.Country, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/countries/GetAll")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxSQLTableNameKey, tableName)

	var result []*countryModel.Country
	query := `SELECT code, name, name_id, created_at, updated_at, deleted_at FROM ` + tableName + ` WHERE deleted_at IS NULL`
	if filter.Name != "" {
		query += ` AND name LIKE '%` + filter.Name + `%'`
	}
	if filter.NameID != "" {
		query += ` AND name_id LIKE '%` + filter.NameID + `%'`
	}
	err := r.db.SelectContext(ctx, &result, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, nil
		}

		r.logger.Error(ctx, "error when get countries", logger.Error(err), logger.String("query", query))
		return nil, err
	}
	return result, nil
}

func (r *countryRepository) FindByCode(ctx context.Context, code string) (*countryModel.Country, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/countries/FindByCode")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxSQLTableNameKey, tableName)

	var result countryModel.Country
	query := `SELECT code, name, name_id, created_at, updated_at, deleted_at FROM ` + tableName + ` WHERE code = ? AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &result, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		r.logger.Error(ctx, "error when get country detail", logger.Error(err), logger.String("query", query), logger.String("code", code))
		return nil, err
	}
	return &result, nil
}
