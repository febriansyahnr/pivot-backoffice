package menuService

import (
	"context"
	"time"

	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
)

// Update is used to update the menu and set updated_at value field only
func (s *MenuService) Update(ctx context.Context, menu *menuModel.Menu) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/menu/Update")
	defer segment.End()

	menu.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, menu)
}

// IsShouldUpdate is used to check whether the existing menu should be updated or not
// currently, it only checks the name and icon
func (s *MenuService) IsShouldUpdate(ctx context.Context, existingMenu *menuModel.Menu, newMenu roleMenuPermissionModel.MenuPermissionFromFileRequest) bool {
	return existingMenu.Name != newMenu.Name || existingMenu.Icon != newMenu.Icon
}
