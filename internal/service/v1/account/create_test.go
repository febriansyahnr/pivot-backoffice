package accountService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateMerchantAccount(t *testing.T) {

	testCases := []struct {
		Name      string
		UserType  string
		MockSetup func(mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository)
		WantErr   bool
	}{
		{
			Name:     "Success Create Merchant Account",
			UserType: constant.UserTypeMerchant,
			WantErr:  false,
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:     "Success Create SubMerchant Account",
			UserType: constant.UserTypeSubMerchant,
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name:     "Error Create Account",
			UserType: constant.UserTypeMerchant,
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIAccountRepository(t)
			trxRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLog, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			ctx := context.Background()
			tc.MockSetup(mockRepo, trxRepo)

			svc := New(mockLog, trxRepo, mockRepo, nil)
			err := svc.CreateMerchantAccount(ctx, "merch-id", tc.UserType)
			if tc.WantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			trxRepo.AssertExpectations(t)
		})
	}
}

func TestCreateAccount(t *testing.T) {

	testCases := []struct {
		Name      string
		Request   *account_model.NewAccountRequest
		Response  *account_model.AccountResponse
		MockSetup func(mockRepo *repositoryMocks.IAccountRepository,
			trxRepo *repositoryMocks.IAccountTransactionRepository,

			merchantSvc *serviceMocks.IMerchantService,
		)
		WantErr bool
	}{
		{
			Name: "SUCCESS: Create Merchant Disbursement Account",
			Request: &account_model.NewAccountRequest{
				ReferenceID: uuid.Max,
				UserType:    constant.UserTypeMerchant,
				Usecase:     constant.ReferenceDisbursement,
				Currency:    constant.CurrencyIDR,
			},
			Response: &account_model.AccountResponse{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceDisbursement,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
				Type:        constant.TypeGeneralLedger,
			},
			WantErr: false,
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository,
				trxRepo *repositoryMocks.IAccountTransactionRepository,

				merchantSvc *serviceMocks.IMerchantService,
			) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "SUCCESS: Create Merchant Payment Account",
			Request: &account_model.NewAccountRequest{
				ReferenceID: uuid.Max,
				UserType:    constant.UserTypeMerchant,
				Usecase:     constant.ReferencePayment,
				Currency:    constant.CurrencyIDR,
			},
			Response: &account_model.AccountResponse{
				ReferenceID: uuid.Max,
				Name:        constant.ReferencePayment,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
				Type:        constant.TypeGeneralLedger,
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository,
				trxRepo *repositoryMocks.IAccountTransactionRepository,

				merchantSvc *serviceMocks.IMerchantService,
			) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Merchant Wallet Account",
			Request: &account_model.NewAccountRequest{
				ReferenceID: uuid.Max,
				UserType:    constant.UserTypeMerchant,
				Usecase:     constant.ReferenceWallet,
				Currency:    constant.CurrencyIDR,
			},
			Response: &account_model.AccountResponse{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceWallet,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
				Type:        constant.TypeGeneralLedger,
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository,
				trxRepo *repositoryMocks.IAccountTransactionRepository,

				merchantSvc *serviceMocks.IMerchantService,
			) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create SubMerchant Disbursement Account",
			Request: &account_model.NewAccountRequest{
				ReferenceID: uuid.Max,
				UserType:    constant.UserTypeSubMerchant,
				Usecase:     constant.ReferenceDisbursement,
				Currency:    constant.CurrencyIDR,
			},
			Response: &account_model.AccountResponse{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceDisbursement,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
				Type:        constant.TypeLedger,
			},
			WantErr: false,
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository,
				trxRepo *repositoryMocks.IAccountTransactionRepository,

				merchantSvc *serviceMocks.IMerchantService,
			) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name: "SUCCESS: Create SubMerchant Payment Account",
			Request: &account_model.NewAccountRequest{
				ReferenceID: uuid.Max,
				UserType:    constant.UserTypeSubMerchant,
				Usecase:     constant.ReferencePayment,
				Currency:    constant.CurrencyIDR,
			},
			Response: &account_model.AccountResponse{
				ReferenceID: uuid.Max,
				Name:        constant.ReferencePayment,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
				Type:        constant.TypeLedger,
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository,
				trxRepo *repositoryMocks.IAccountTransactionRepository,

				merchantSvc *serviceMocks.IMerchantService,
			) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create SubMerchant Wallet Account",
			Request: &account_model.NewAccountRequest{
				ReferenceID: uuid.Max,
				UserType:    constant.UserTypeSubMerchant,
				Usecase:     constant.ReferenceWallet,
				Currency:    constant.CurrencyIDR,
			},
			Response: &account_model.AccountResponse{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceWallet,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
				Type:        constant.TypeLedger,
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository,
				trxRepo *repositoryMocks.IAccountTransactionRepository,

				merchantSvc *serviceMocks.IMerchantService,
			) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Customer User Type",
			Request: &account_model.NewAccountRequest{
				ReferenceID: uuid.Max,
				UserType:    constant.UserTypeCustomer,
				Usecase:     constant.ReferenceWallet,
				Currency:    constant.CurrencyIDR,
			},
			Response: &account_model.AccountResponse{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceWallet,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeCustomer,
				Type:        constant.TypeLedger,
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository,
				trxRepo *repositoryMocks.IAccountTransactionRepository,

				merchantSvc *serviceMocks.IMerchantService,
			) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Create Account",
			Request: &account_model.NewAccountRequest{
				ReferenceID: uuid.Max,
				UserType:    constant.UserTypeMerchant,
				Usecase:     constant.ReferenceWallet,
				Currency:    constant.CurrencyIDR,
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository,
				trxRepo *repositoryMocks.IAccountTransactionRepository,

				merchantSvc *serviceMocks.IMerchantService,
			) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIAccountRepository(t)
			trxRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLog, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			merchantSvc := serviceMocks.NewIMerchantService(t)
			ctx := context.Background()
			tc.MockSetup(mockRepo, trxRepo, merchantSvc)

			svc := New(mockLog, trxRepo, mockRepo, nil)
			resp, err := svc.CreateAccount(ctx, tc.Request)
			if tc.WantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.UUID)
				assert.NotEmpty(t, resp.CreatedAt)
				assert.NotEmpty(t, resp.UpdatedAt)
				assert.NotEmpty(t, resp.LastUpdateBalanceAt)
				assert.Empty(t, resp.CurrentBalanceCheckTime)
				assert.Equal(t, tc.Response.Name, resp.Name)
				assert.Equal(t, tc.Response.Currency, resp.Currency)
				assert.Equal(t, tc.Response.CurrentBalance, resp.CurrentBalance)
				assert.Equal(t, tc.Response.EODBalance, resp.EODBalance)
				assert.Equal(t, tc.Response.ReferenceID, resp.ReferenceID)
				assert.Equal(t, tc.Response.Type, resp.Type)
				assert.Equal(t, tc.Response.UserType, resp.UserType)
			}

			mockRepo.AssertExpectations(t)
			trxRepo.AssertExpectations(t)

		})
	}

}
