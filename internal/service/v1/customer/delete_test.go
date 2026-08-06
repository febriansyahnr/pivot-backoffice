package customerService

import (
	"context"
	"fmt"
	"testing"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	logger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	mockLogger, _ := logger.NewZapLogger(logger.Config{})

	testCases := []struct {
		name       string
		setup      func(customerRepo *repositoryMock.ICustomerRepository)
		merchantId string
		id         string
		wantErr    bool
	}{
		{
			name: "SUCCESS: Delete customer",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, "customer-id", "parent-merchant-id").Return(
					&customerModel.Customer{
						UUID:       "customer-id",
						MerchantID: "parent-merchant-id",
					}, nil)
				customerRepo.On("Delete", mock.Anything, "customer-id", "parent-merchant-id").Return(nil)
			},
			merchantId: "parent-merchant-id",
			id:         "customer-id",
		},
		{
			name: "FAILED: Failed when delete customer, not found",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, "customer-id", "non-parent-merchant-id").Return(
					nil, nil)
			},
			merchantId: "non-parent-merchant-id",
			id:         "customer-id",
			wantErr:    true,
		},
		{
			name: "FAILED: Failed when delete customer, unauthorized",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, "customer-id", "non-parent-merchant-id").Return(
					&customerModel.Customer{
						UUID:       "customer-id",
						MerchantID: "non-parent-merchant-id",
					}, nil)
				customerRepo.On("Delete", mock.Anything, "customer-id", "non-parent-merchant-id").Return(fmt.Errorf("Delete error"))
			},
			merchantId: "non-parent-merchant-id",
			id:         "customer-id",
			wantErr:    true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo)
			service := New(customerRepo, accountService, mockLogger)
			_, err := service.DeleteCustomer(context.Background(), tc.id, tc.merchantId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
