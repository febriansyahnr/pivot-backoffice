package location

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/location"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

func (r *repository) GetProvinces(ctx context.Context) (resp []location.Province, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/location/GetProvinces")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, provinceTableName)

	err = r.db.SelectContext(ctx, &resp, `SELECT id, name FROM `+provinceTableName+` ORDER BY name`)
	return
}

func (r *repository) GetCitiesByProvinceId(ctx context.Context, provinceId uint16) (resp []location.City, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/location/GetCitiesByProvinceId")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, cityTableName)

	rawQuery := `SELECT id, name FROM ` + cityTableName + ` WHERE province_id = ? ORDER BY name`

	err = r.db.SelectContext(ctx, &resp, rawQuery, provinceId)
	return
}

func (r *repository) GetDistrictsByCityId(ctx context.Context, cityId uint16) (resp []location.District, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/location/GetDistrictsByCityId")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, districtTableName)

	rawQuery := `SELECT id, name FROM ` + districtTableName + ` WHERE city_id = ? ORDER BY name`

	err = r.db.SelectContext(ctx, &resp, rawQuery, cityId)
	return
}

func (r *repository) GetDistrictById(ctx context.Context, id uint16) (resp *location.District, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/location/GetDistrictById")
	defer segment.End()

	ctx, resp = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, districtTableName), &location.District{}

	err = r.db.GetContext(
		ctx, resp, `SELECT id, city_id, name FROM `+districtTableName+` WHERE id = ?;`, id,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return
}
