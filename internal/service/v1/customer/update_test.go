package customerService

import (
	"context"
	"errors"
	"testing"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	logger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateCustomer(t *testing.T) {
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	testCases := []struct {
		name    string
		setup   func(customerRepo *repositoryMock.ICustomerRepository)
		request customerModel.UpdateCustomerRequest
		wantErr bool
	}{
		{
			name: "SUCCESS: Update customer",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, mock.Anything, mock.Anything).Return(&customerModel.Customer{
					UUID:       "123",
					FirstName:  "John",
					LastName:   "Doe",
					MerchantID: "123",
				}, nil)
				customerRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			request: customerModel.UpdateCustomerRequest{
				UUID:       "123",
				FirstName:  customerModel.NewString("John"),
				LastName:   customerModel.NewString("Doe"),
				MerchantID: "123",
			},
		},
		{
			name: "ERROR: Failed when update customer",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, mock.Anything, mock.Anything).Return(&customerModel.Customer{
					UUID:       "123",
					FirstName:  "John",
					LastName:   "Doe",
					MerchantID: "123",
				}, nil)
				customerRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("Failed to update customer"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Phone number already exist",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, mock.Anything, mock.Anything).Return(&customerModel.Customer{
					UUID:       "123",
					FirstName:  "John",
					LastName:   "Doe",
					MerchantID: "123",
				}, nil)
				customerRepo.On("GetCustomerByPhoneNumber", mock.Anything, mock.Anything, mock.Anything).Return(&customerModel.Customer{
					UUID:      "456",
					FirstName: "Stephen",
					LastName:  "Doe",
				}, nil)
			},
			request: customerModel.UpdateCustomerRequest{
				PhoneNumber: customerModel.NewString("081234567890"),
			},
			wantErr: true,
		},
		{
			name: "ERROR: Failed customer not found",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On("GetCustomerById", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("Customer not found"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo)

			service := New(customerRepo, accountService, mockLogger)
			_, err := service.UpdateCustomer(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
