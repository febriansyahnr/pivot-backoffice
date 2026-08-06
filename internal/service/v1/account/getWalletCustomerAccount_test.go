package accountService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	svcMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWalletCustomerAccount(t *testing.T) {

	testcases := []struct {
		Name      string
		MockSetup func(mockRepo *repositoryMocks.IAccountRepository, customerSvc *svcMocks.ICustomerService)
		WantErr   bool
	}{
		{
			Name: "SUCCESS: Get Wallet Customer Account",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, customerSvc *svcMocks.ICustomerService) {
				customerSvc.On(
					"GetCustomerById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(
					&customerModel.GeneralCustomerResponse{},
					nil,
				)

				mockRepo.On(
					"GetByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&account_model.Account{}, nil)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Get Customer",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, customerSvc *svcMocks.ICustomerService) {
				customerSvc.On(
					"GetCustomerById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(
					nil,
					errors.New("error"),
				)
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Get By Reference ID and Usecase",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, customerSvc *svcMocks.ICustomerService) {
				customerSvc.On(
					"GetCustomerById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(
					&customerModel.GeneralCustomerResponse{},
					nil,
				)

				mockRepo.On(
					"GetByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))

			},
			WantErr: true,
		},
		{
			Name: "ERROR: Account Not found",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, customerSvc *svcMocks.ICustomerService) {
				customerSvc.On(
					"GetCustomerById",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(
					&customerModel.GeneralCustomerResponse{},
					nil,
				)

				mockRepo.On(
					"GetByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)

			},
			WantErr: true,
		},
	}
	for _, tc := range testcases {

		t.Run(tc.Name, func(t *testing.T) {
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewIAccountRepository(t)
			customerSvc := svcMocks.NewICustomerService(t)
			ctx := context.Background()

			tc.MockSetup(mockRepo, customerSvc)
			svc := New(loggerMock, nil, mockRepo, nil)
			WithCustomerService(svc, customerSvc)

			account, err := svc.GetWalletCustomerAccount(ctx, &account_model.GetCustomerAccountRequest{
				CustomerID: uuid.NewString(),
				MerchantID: uuid.NewString(),
			})

			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, account)
			}

		})
	}

}
