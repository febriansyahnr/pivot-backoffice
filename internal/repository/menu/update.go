package menuRepository

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Update is used to update the name, icon, and updated_at field only
// from the menu table based on the uuid
func (r *MenuRepository) Update(ctx context.Context, menu *menuModel.Menu) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/menu/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, mySqlExt.CtxSQLTableNameKey, "menus")

	query := `UPDATE menus 
		SET name = :name, 
			icon = :icon, 
			updated_at = :updated_at 
		WHERE uuid = :uuid`

	affected, err := r.db.NamedExecContext(ctx, query, menu)
	if err != nil {
		r.logger.Error(ctx, "error when update the menu", logger.Error(err))
		return err
	}

	if !affected {
		r.logger.Warn(ctx, "menu information is not updated", logger.String("menuId", menu.UUID))
		return constant.ErrNoRowsAffected
	}

	return nil
}
