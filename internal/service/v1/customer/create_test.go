package customerService

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	logger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateCustomer(t *testing.T) {
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	testCases := []struct {
		name    string
		setup   func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService)
		request customerModel.CreateCustomerRequest
		wantErr bool
	}{
		{
			name: "SUCCESS: Create customer and wallet",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetCustomerByPhoneNumber", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				customerRepo.On("Create", mock.Anything, mock.Anything).Return(nil, nil)
				accountService.On("CreateAccount", mock.Anything, mock.Anything).Return(nil, nil)
			},
			request: customerModel.CreateCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
			},
		},
		{
			name: "ERROR: Customer already created",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetCustomerByPhoneNumber",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(&customerModel.Customer{
					UUID: uuid.NewString(),
				}, nil)
			},
			request: customerModel.CreateCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Failed when create customer",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetCustomerByPhoneNumber", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				customerRepo.On("Create", mock.Anything, mock.Anything).Return(fmt.Errorf("Create error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Failed when create account",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetCustomerByPhoneNumber", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				customerRepo.On("Create", mock.Anything, mock.Anything).Return(nil, nil)
				accountService.On("CreateAccount", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Create account error"))
			},
			request: customerModel.CreateCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Error when get customer with phone number",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetCustomerByPhoneNumber", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("get customer error"))
			},
			request: customerModel.CreateCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Phone number already exist",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetCustomerByPhoneNumber", mock.Anything, mock.Anything, mock.Anything).Return(&customerModel.Customer{}, nil)
			},
			request: customerModel.CreateCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo, accountService)

			service := New(customerRepo, accountService, mockLogger)
			_, err := service.CreateCustomer(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateUnfiedPaymentCustomer(t *testing.T) {
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	testCases := []struct {
		name    string
		setup   func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService)
		request customerModel.CreateUnifiedPaymentCustomerRequest
		wantErr bool
	}{
		{
			name: "SUCCESS: Create customer",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetMerchantCustomerByEmail",
					constant.ValueCtxMockType(),
					mock.Anything).Return(nil, nil).Once()
				customerRepo.On("Create", mock.Anything, mock.Anything).Return(nil, nil).Once()
			},
			request: customerModel.CreateUnifiedPaymentCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
			},
		},
		{
			name: "SUCCESS: when customer already exist with metadata, then should merge and update it",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				existingMetadata := map[string]interface{}{
					"existingKey": "existingValue",
					"sharedKey":   "oldValue",
				}
				customerRepo.On("GetMerchantCustomerByEmail",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&customerModel.Customer{
					UUID:     uuid.NewString(),
					Metadata: existingMetadata,
				}, nil)
				customerRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			request: customerModel.CreateUnifiedPaymentCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
				Email:      "john@example.com",
				Metadata: map[string]interface{}{
					"newKey":    "newValue",
					"sharedKey": "newValue",
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: when customer already exist with metadata but request has no metadata, then should merge empty map",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				existingMetadata := map[string]interface{}{
					"existingKey": "existingValue",
				}
				customerRepo.On("GetMerchantCustomerByEmail",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&customerModel.Customer{
					UUID:     uuid.NewString(),
					Metadata: existingMetadata,
				}, nil)
				customerRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			request: customerModel.CreateUnifiedPaymentCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
				Email:      "john@example.com",
				Metadata:   nil, // No metadata provided
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: when customer already exist without metadata, then should skip merge and update",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetMerchantCustomerByEmail",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&customerModel.Customer{
					UUID:     uuid.NewString(),
					Metadata: nil, // No existing metadata
				}, nil)
				customerRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			request: customerModel.CreateUnifiedPaymentCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
				Email:      "john@example.com",
				Metadata: map[string]interface{}{
					"newKey": "newValue",
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failed when create customer, then should return error",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetMerchantCustomerByEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				customerRepo.On("Create", mock.Anything, mock.Anything).Return(fmt.Errorf("Create error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Error when get customer with email",
			setup: func(customerRepo *repositoryMock.ICustomerRepository, accountService *serviceMock.IAccountService) {
				customerRepo.On("GetMerchantCustomerByEmail", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			request: customerModel.CreateUnifiedPaymentCustomerRequest{
				FirstName:  "John",
				LastName:   "Doe",
				MerchantID: "123",
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo, accountService)

			service := New(customerRepo, accountService, mockLogger)
			_, err := service.CreateUnfiedPaymentCustomer(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
