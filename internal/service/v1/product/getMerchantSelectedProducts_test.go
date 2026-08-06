package productService

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
)

func TestGetMerchantSelectedProducts(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(productRepo *mockRepo.IProductRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS: Get merchant selected products",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetMerchantSelectedProducts",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return([]*product.MerchantWithProductName{}, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Get merchant selected products",
			setup: func(productRepo *mockRepo.IProductRepository) {
				productRepo.On(
					"GetMerchantSelectedProducts",
					constant.ValueCtxMockType(),
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
			products, err := svc.GetMerchantSelectedProducts(context.Background(), uuid.NewString())

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
func TestGetMerchantActiveProducts(t *testing.T) {
	productRepo := mockRepo.NewIProductRepository(t)
	merchantRepo := mockRepo.NewIMerchantRepository(t)
	logger, _ := logger.NewZapLogger(logger.Config{})

	svc := New(logger, productRepo, WithMerchantRepo(merchantRepo))

	products := []*product.MerchantWithProductName{
		{
			ProductID:   uuid.NewString(),
			ProductName: "PLATFORM",
			Active:      true,
		},
	}

	testCases := []struct {
		name          string
		setup         func(productRepo *mockRepo.IProductRepository, merchantRepo *mockRepo.IMerchantRepository)
		shouldErr     bool
		want          []*product.MerchantWithProductName
		expectedError error
	}{
		{
			name: "SUCCESS: Get merchant active products",
			setup: func(productRepo *mockRepo.IProductRepository, merchantRepo *mockRepo.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&merchant.Merchant{
					KYCStatus: sql.NullString{
						String: constant.KYCStatusApproved,
						Valid:  true,
					},
				}, nil).Once()

				productRepo.On(
					"GetMerchantActiveProducts",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(products, nil).Once()
			},
			shouldErr: false,
			want:      products,
		},
		{
			name: "Error: error get merchant",
			setup: func(productRepo *mockRepo.IProductRepository, merchantRepo *mockRepo.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr:     true,
			expectedError: pkgErrors.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR: Merchant not found",
			setup: func(productRepo *mockRepo.IProductRepository, merchantRepo *mockRepo.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(nil, nil).Once()
			},
			shouldErr:     true,
			expectedError: pkgErrors.New(response.HttpErrNotFound, constant.ErrMerchantNotFound),
		},
		{
			name: "ERROR: Forbidden KYC status",
			setup: func(productRepo *mockRepo.IProductRepository, merchantRepo *mockRepo.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&merchant.Merchant{KYCStatus: sql.NullString{
					String: constant.KYCStatusInReview,
					Valid:  true,
				}}, nil).Once()

				productRepo.On(
					"GetMerchantActiveProducts",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(products, nil).Once()
			},
			shouldErr:     true,
			expectedError: pkgErrors.New(response.HttpErrForbidden, constant.ErrForbiddenKYCStatusAccess),
		},
		{
			name: "SUCCESS: Get merchant active products with KYC status not required",
			setup: func(productRepo *mockRepo.IProductRepository, merchantRepo *mockRepo.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&merchant.Merchant{
					KYCStatus: sql.NullString{
						String: constant.KYCStatusNotRequired,
						Valid:  true,
					},
				}, nil).Once()

				productRepo.On(
					"GetMerchantActiveProducts",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(products, nil).Once()
			},
			shouldErr: false,
			want:      []*product.MerchantWithProductName{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(productRepo, merchantRepo)
			products, err := svc.GetMerchantActiveProducts(context.Background(), uuid.NewString())

			if tc.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
				return
			}

			assert.Nil(t, err)
			assert.NotNil(t, products)
			assert.Equal(t, tc.want, products)
		})
	}
}
