package productService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
)

func (s *ProductService) GetMerchantSelectedProducts(ctx context.Context, merchantId string) ([]*product.MerchantWithProductName, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/product/GetMerchantSelectedProducts")
	defer segment.End()

	products, err := s.repo.GetMerchantSelectedProducts(ctx, merchantId)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetMerchantSelectedProducts)
	}

	return products, nil
}

// GetMerchantActiveProducts retrieves a list of active products for a given merchant.
func (s *ProductService) GetMerchantActiveProducts(ctx context.Context, merchantID string) ([]*product.MerchantWithProductName, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/product/GetMerchantSelectedProducts")
	defer segment.End()

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, err)
	}

	if merchant == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	products, err := s.repo.GetMerchantActiveProducts(ctx, merchantID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrGetMerchantSelectedProducts)
	}

	if merchant.KYCStatus.String == constant.KYCStatusApproved {
		return products, nil
	}

	if merchant.KYCStatus.String != constant.KYCStatusNotRequired {
		return nil, pkgErrors.New(response.HttpErrForbidden, constant.ErrForbiddenKYCStatusAccess)
	}

	activeProduct := []*product.MerchantWithProductName{}
	for _, product := range products {
		// when the product is restricted, then skip it
		if _, ok := restrictedNonKYCProduct[product.ProductName]; ok {
			continue
		}

		activeProduct = append(activeProduct, product)
	}

	return activeProduct, nil
}
