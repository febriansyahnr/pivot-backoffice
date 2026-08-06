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

func TestProductList(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(productRepo *mockRepo.IProductRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS: Get product list",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetProductList",
					constant.ValueCtxMockType(),
				).Return([]*product.Product{}, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Get product list",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetProductList",
					constant.ValueCtxMockType(),
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
			products, err := svc.GetProductList(context.Background())

			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, products)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, products)
			}
		})
	}

}
