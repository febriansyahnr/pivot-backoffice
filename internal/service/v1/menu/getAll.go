package menuService

import (
	"context"

	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
)

func (s *MenuService) GetAll(ctx context.Context, excludeHome bool) ([]*menuModel.MenuResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/menu/GetAll")
	defer segment.End()

	menus, err := s.repo.GetAll(ctx, &menuModel.GetAllFilterRequest{})
	if err != nil {
		return nil, err
	}

	// Filter out Home menu if requested (used for role management UI)
	if excludeHome {
		filtered := make([]*menuModel.MenuResponse, 0, len(menus))
		for _, menu := range menus {
			if menu.Slug != "home" {
				filtered = append(filtered, menu)
			}
		}
		return filtered, nil
	}

	return menus, nil
}
