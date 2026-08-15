package vendor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *VendorRepository) GetByID(ctx context.Context, id string) (*vendorModel.Vendor, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/vendor/GetByID")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := fmt.Sprintf(`SELECT %s FROM %s v WHERE v.uuid = ? AND v.deleted_at IS NULL`, tableColumns, tableName)

	var vendor vendorModel.Vendor
	err := r.db.GetContext(ctx, &vendor, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when getting vendor by id", logger.Error(err), logger.String("id", id))
		return nil, err
	}

	return &vendor, nil
}

func (r *VendorRepository) List(ctx context.Context, q *vendorModel.VendorQuery) ([]*vendorModel.Vendor, int, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/vendor/List")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	whereClause, args := q.BuildCondition()
	if whereClause != "" {
		whereClause = "WHERE " + whereClause
	}

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s v %s`, tableName, whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error(ctx, "error when counting vendors", logger.Error(err))
		return nil, 0, err
	}

	// Data query with pagination
	offset := (q.Page - 1) * q.PageSize
	orderBy := q.BuildOrderBy()
	dataQuery := fmt.Sprintf(`SELECT %s FROM %s v %s ORDER BY v.%s LIMIT ? OFFSET ?`, tableColumns, tableName, whereClause, orderBy)

	// Append pagination args
	args = append(args, q.PageSize, offset)

	var vendors []*vendorModel.Vendor
	err = r.db.SelectContext(ctx, &vendors, dataQuery, args...)
	if err != nil {
		r.logger.Error(ctx, "error when listing vendors", logger.Error(err))
		return nil, 0, err
	}

	return vendors, total, nil
}

func (r *VendorRepository) GetByName(ctx context.Context, merchantID, name string) (*vendorModel.Vendor, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/vendor/GetByName")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxSQLTableNameKey, tableName)

	query := fmt.Sprintf(`SELECT %s FROM %s v WHERE v.merchant_id = ? AND v.name = ? AND v.deleted_at IS NULL`, tableColumns, tableName)

	var vendor vendorModel.Vendor
	err := r.db.GetContext(ctx, &vendor, query, merchantID, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error(ctx, "error when getting vendor by name", logger.Error(err), logger.String("merchantID", merchantID), logger.String("name", name))
		return nil, err
	}

	return &vendor, nil
}
