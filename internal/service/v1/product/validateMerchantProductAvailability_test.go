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
)

func TestValidateMerchantProductAvailability(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(productRepo *mockRepo.IProductRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS: Merchant activated product",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetMerchantSelectedProductByName",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&product.MerchantWithProductName{
					Active: true,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "FAILED: Merchant not activated product",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetMerchantSelectedProductByName",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&product.MerchantWithProductName{}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Get merchant selected products",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetMerchantSelectedProductByName",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil, errors.New("errors"))
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
			err := svc.ValidateMerchantProductAvailability(context.Background(), &product.ValidateMerchantProductAvailability{})

			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
