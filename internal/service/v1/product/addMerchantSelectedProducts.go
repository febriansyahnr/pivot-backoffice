package productService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *ProductService) AddMerchantSelectedProduct(ctx context.Context, request *product.AddMerchantProductRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/product/AddMerchantSelectedProduct")
	defer segment.End()

	existingProduct, err := s.repo.GetProductById(ctx, request.ProductID)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrGetProduct)
	}
	if existingProduct == nil {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrProductNotFound)
	}

	existing, err := s.repo.GetMerchantSelectedProductById(ctx, request.MerchantID, request.ProductID)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrGetExistingMerchantSelectedProduct)
	}
	if existing != nil {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrMerchantSelectedProductAlreadyExists)
	}

	data := product.NewMerchantSelectedProduct(request)
	err = s.repo.AddMerchantSelectedProduct(ctx, data)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrAddMerchantSelectedProduct)
	}

	return nil
}
