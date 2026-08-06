package productService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
)

func (s *ProductService) GetProductList(ctx context.Context) ([]*product.Product, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/product/GetProductList")
	defer segment.End()

	products, err := s.repo.GetProductList(ctx)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetProductList)
	}

	return products, nil
}
