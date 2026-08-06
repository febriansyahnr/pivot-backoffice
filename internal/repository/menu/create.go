package menuRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *MenuRepository) Create(ctx context.Context, menu *menuModel.Menu) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/menu/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "menus")

	query := "INSERT INTO menus (uuid, slug, name, type, icon, path, level, `order`, parent_id, created_at, updated_at) " +
		"VALUES (:uuid, :slug, :name, :type, :icon, :path, :level, :order, :parent_id, :created_at, :updated_at)"

	affected, err := r.db.NamedExecContext(ctx, query, menu)
	if err != nil {
		r.logger.Error(ctx, "error when inserting menu", logger.Error(err))
		return err
	}

	if !affected {
		err = constant.ErrNoRowsAffected
		r.logger.Error(ctx, "failed when inserting menu", logger.Error(err))
		return err
	}

	return nil
}
