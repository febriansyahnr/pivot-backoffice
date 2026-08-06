package productService

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateMerchantProductAvailability(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(productRepo *mockRepo.IProductRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS: Update merchant product availablity",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					&product.Product{},
					nil,
				)

				productRepo.On(
					"GetMerchantSelectedProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&product.MerchantWithProductName{}, nil)

				productRepo.On(
					"UpdateMerchantProductAvailability",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*product.UpdateMerchantSelectedProductAvailabilityRequest"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Find product",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					nil,
					errors.New("errors"),
				)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Product not found",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					nil,
					nil,
				)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Error get Merchant selected product",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					&product.Product{},
					nil,
				)

				productRepo.On(
					"GetMerchantSelectedProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil, errors.New("errors"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Merchant not selected product",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					&product.Product{},
					nil,
				)

				productRepo.On(
					"GetMerchantSelectedProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Update merchant product availability",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(
					&product.Product{},
					nil,
				)

				productRepo.On(
					"GetMerchantSelectedProductById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&product.MerchantWithProductName{}, nil)

				productRepo.On(
					"UpdateMerchantProductAvailability",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*product.UpdateMerchantSelectedProductAvailabilityRequest"),
				).Return(errors.New("error"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			productRepo := mockRepo.NewIProductRepository(t)
			logger, _ := logger.NewZapLogger(logger.Config{})
			tc.setup(productRepo)

			svc := New(logger, productRepo)
			err := svc.UpdateMerchantProductAvailability(context.Background(), &product.UpdateMerchantSelectedProductAvailabilityRequest{})

			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
