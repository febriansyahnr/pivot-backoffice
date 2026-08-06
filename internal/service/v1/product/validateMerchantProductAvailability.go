package productService

import (
	"context"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *ProductService) ValidateMerchantProductAvailability(ctx context.Context, request *product.ValidateMerchantProductAvailability) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/product/ValidateMerchantProductAvailability")
	defer segment.End()

	selectedProduct, err := s.repo.GetMerchantSelectedProductByName(ctx, request.MerchantID, request.ProductName)
	if err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrValidateMerchantProductAvailability)
	}
	if selectedProduct == nil || !selectedProduct.Active {
		errMsg := fmt.Sprintf(constant.MerchantIsNotAllowedToUseProductMsgFormat, request.ProductName)
		return pkgErrors.New(response.HttpErrForbidden, errors.New(errMsg))
	}

	return nil
}
