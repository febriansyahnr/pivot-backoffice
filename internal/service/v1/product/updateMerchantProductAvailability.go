package productService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
)

func (s *ProductService) UpdateMerchantProductAvailability(ctx context.Context, request *product.UpdateMerchantSelectedProductAvailabilityRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/product/UpdateMerchantProductAvailability")
	defer segment.End()

	existingProduct, err := s.repo.GetProductById(ctx, request.ProductID)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrGetProduct)
	}
	if existingProduct == nil {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrProductNotFound)
	}

	selectedProduct, err := s.repo.GetMerchantSelectedProductById(ctx, request.MerchantID, request.ProductID)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrGetExistingMerchantSelectedProduct)
	}
	if selectedProduct == nil {
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrMerchantNotSelectedProduct)
	}

	err = s.repo.UpdateMerchantProductAvailability(ctx, request)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrUpdateMerchantProductAvailability)
	}

	return nil
}
