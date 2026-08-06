package accountService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantAccounts(t *testing.T) {
	sourceMerchantID := uuid.New()
	recipientMerchantID := uuid.New()
	testCases := []struct {
		name      string
		input     []uuid.UUID
		setup     func(accountRepo *mockRepo.IAccountRepository)
		expectErr bool
	}{
		{
			name: "SUCCESS: Get Merchant Accounts",
			setup: func(accountRepo *mockRepo.IAccountRepository) {
				accountRepo.On(
					"GetEntityAccounts",
					mock.Anything,
					mock.AnythingOfType("[]uuid.UUID"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(
					map[uuid.UUID]*account_model.Account{
						sourceMerchantID: {
							UUID: uuid.New(),
						},
						recipientMerchantID: {
							UUID: uuid.New(),
						},
					},
					nil,
				)
			},
			input: []uuid.UUID{sourceMerchantID, recipientMerchantID},
		},
		{
			name: "SUCCESS: Get Merchant Accounts with create new one",
			setup: func(accountRepo *mockRepo.IAccountRepository) {
				accountRepo.On(
					"GetEntityAccounts",
					mock.Anything,
					mock.AnythingOfType("[]uuid.UUID"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(
					map[uuid.UUID]*account_model.Account{
						sourceMerchantID: {
							UUID: uuid.New(),
						},
					},
					nil,
				)

				accountRepo.On(
					"Create",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(nil)
			},
			input: []uuid.UUID{sourceMerchantID, recipientMerchantID},
		},
		{
			name: "ERROR: Empty merchant ids",
			setup: func(accountRepo *mockRepo.IAccountRepository) {

			},
			input:     []uuid.UUID{},
			expectErr: false,
		},
	}
	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			accountRepo := mockRepo.NewIAccountRepository(t)
			tc.setup(accountRepo)

			svc := New(nil, nil, accountRepo, nil)
			resp, err := svc.GetMerchantAccounts(context.TODO(), tc.input, "")
			if tc.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				if len(tc.input) > 0 {
					assert.NotNil(t, resp)
				} else {
					assert.Nil(t, resp)
				}
			}
		})
	}
}

func TestGetWalletMerchantAccount(t *testing.T) {

	merchantId := uuid.New()
	parentMerchantId := uuid.New()

	testcases := []struct {
		Name      string
		MockSetup func(mockRepo *repositoryMocks.IAccountRepository, merchantSvc *serviceMocks.IMerchantService)
		WantErr   bool
	}{
		{
			Name: "SUCCESS: Get By Reference ID and Usecase",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, merchantSvc *serviceMocks.IMerchantService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
			Name: "ERROR: Incorrect parent merchant",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, merchantSvc *serviceMocks.IMerchantService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Account not found",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, merchantSvc *serviceMocks.IMerchantService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
		{
			Name: "ERROR: Get By Reference ID and Usecase",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, merchantSvc *serviceMocks.IMerchantService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
	}
	for _, tc := range testcases {

		t.Run(tc.Name, func(t *testing.T) {
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewIAccountRepository(t)
			merchantSvc := serviceMocks.NewIMerchantService(t)
			ctx := context.Background()

			tc.MockSetup(mockRepo, merchantSvc)
			svc := New(loggerMock, nil, mockRepo, nil)
			WithMerchantService(svc, merchantSvc)
			account, err := svc.GetWalletMerchantAccount(ctx, parentMerchantId, merchantId)

			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, account)
			}

		})
	}
}
