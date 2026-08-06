package menuService

import (
	"context"

	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
)

func (s *MenuService) Create(ctx context.Context, menu *menuModel.Menu) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/menu/Create")
	defer segment.End()

	return s.repo.Create(ctx, menu)
}
