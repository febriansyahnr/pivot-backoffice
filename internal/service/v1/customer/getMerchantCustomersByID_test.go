package customerService

import (
	"context"
	"errors"
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

func TestGetMerchantCustomersByID(t *testing.T) {

	testCases := []struct {
		name    string
		setup   func(customerRepo *repositoryMock.ICustomerRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS: Get merchant customer",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On(
					"GetMerchantCustomersByID",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("[]string"),
				).Return(
					[]*customerModel.Customer{},
					nil,
				)
			},
		},
		{
			name: "ERROR: Error get merchant customer",
			setup: func(customerRepo *repositoryMock.ICustomerRepository) {
				customerRepo.On(
					"GetMerchantCustomersByID",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("[]string"),
				).Return(
					nil,
					errors.New("error"),
				)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := logger.NewZapLogger(logger.Config{})

			customerRepo := repositoryMock.NewICustomerRepository(t)
			accountService := serviceMock.NewIAccountService(t)
			tc.setup(customerRepo)

			service := New(customerRepo, accountService, mockLogger)
			_, err := service.GetMerchantCustomersByID(context.Background(), uuid.NewString(), []string{uuid.NewString()})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
