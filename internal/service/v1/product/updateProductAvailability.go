package productService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *ProductService) UpdateProductAvailability(ctx context.Context, request *product.UpdateProductRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/product/UpdateProductAvailability")
	defer segment.End()

	product, err := s.repo.GetProductById(ctx, request.ID)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrGetProduct)
	}
	if product == nil {
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrProductNotFound)
	}

	err = s.repo.UpdateProductAvailability(ctx, request)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrUpdateProductAvailability)
	}

	return nil

}
