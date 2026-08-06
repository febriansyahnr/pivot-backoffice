package menuService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *MenuService) FindBySlug(ctx context.Context, slug string) (*menuModel.Menu, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/menu/FindBySlug")
	defer segment.End()

	res, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if res == nil {
		err = constant.ErrMenuNotFound
		return nil, pkgErrors.New(response.HttpErrNotFound, err)
	}

	return res, nil
}
