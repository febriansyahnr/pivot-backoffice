package orchestrator_service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAvailableMerchantBalance(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
			balanceRepo *repositoryMocks.IAccountRepository,

		)
		inputMerchantID string
		wantErr         bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(&account_model.Account{}, nil)

				accTrxRepo.On(
					"GetAggregateTransactions",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					mock.AnythingOfType("*orchestrator_model.GetAggregateRequest"),
				).Return(&orchestratorModel.AggregateResponse{}, nil)
			},
			inputMerchantID: uuid.NewString(),
			wantErr:         false,
		},
		{
			name: "SUCCESS: With create balance first",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				balanceRepo.On(
					"Create",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					constant.PtrAccountMockType(),
				).Return(nil)

				accTrxRepo.On(
					"GetAggregateTransactions",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					mock.AnythingOfType("*orchestrator_model.GetAggregateRequest"),
				).Return(&orchestratorModel.AggregateResponse{}, nil)
			},
			inputMerchantID: uuid.NewString(),
			wantErr:         false,
		},
		{
			name: "ERROR: Invalid merchant ID",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,

			) {
			},
			inputMerchantID: "sss",
			wantErr:         true,
		},
		{
			name: "ERROR: FindMerchantAccountByName error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			inputMerchantID: uuid.NewString(),
			wantErr:         true,
		},
		{
			name: "ERROR: GetAggregateTransactions error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(&account_model.Account{}, nil)

				accTrxRepo.On(
					"GetAggregateTransactions",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					mock.AnythingOfType("*orchestrator_model.GetAggregateRequest"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			inputMerchantID: uuid.NewString(),
			wantErr:         true,
		},
		{
			name: "ERROR: Balance create",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

				balanceRepo.On(
					"Create",
					mock.AnythingOfType(constant.MockTypeBackgroundContext),
					constant.PtrAccountMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			inputMerchantID: uuid.NewString(),
			wantErr:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			balanceRepoMock := repositoryMocks.NewIAccountRepository(t)
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock, balanceRepoMock)

			accTrxSvc := New(loggerMock, nil, accTrxRepoMock, balanceRepoMock)
			ctx := context.Background()
			_, err := accTrxSvc.GetAvailableMerchantBalance(ctx, tc.inputMerchantID, constant.TypeDisbursement)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
			balanceRepoMock.AssertExpectations(t)

		})
	}
}

func TestGetMerchantBulkBalances(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, accountSvc *serviceMocks.IAccountService)
		request    *account_model.GetBulkBalanceRequest
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, accountSvc *serviceMocks.IAccountService) {
				merchantID := uuid.New()
				accountUUID := uuid.New()
				accountList := map[uuid.UUID]*account_model.Account{
					merchantID: {UUID: accountUUID, EODBalance: 100.0},
				}
				
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(accountList, nil)
				
				bulkResponse := []*orchestratorModel.BulkAggregateResponse{
					{AccountID: accountUUID.String(), SumOfCredit: 50.0, SumOfDebit: 20.0},
				}
				accTrxRepo.On("GetBulkAggregateTransactions", mock.Anything, mock.Anything).Return(bulkResponse, nil)
			},
			request: &account_model.GetBulkBalanceRequest{
				MerchantIDs: []uuid.UUID{uuid.New()},
				Usecase:     "test_usecase",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Empty account list",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, accountSvc *serviceMocks.IAccountService) {
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(map[uuid.UUID]*account_model.Account{}, nil)
			},
			request: &account_model.GetBulkBalanceRequest{
				MerchantIDs: []uuid.UUID{uuid.New()},
				Usecase:     "test_usecase",
			},
			wantErr: false,
		},
		{
			name: "ERROR: GetMerchantAccounts error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, accountSvc *serviceMocks.IAccountService) {
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			request: &account_model.GetBulkBalanceRequest{
				MerchantIDs: []uuid.UUID{uuid.New()},
				Usecase:     "test_usecase",
			},
			wantErr: true,
		},
		{
			name: "ERROR: GetBulkAggregateTransactions error",
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, accountSvc *serviceMocks.IAccountService) {
				merchantID := uuid.New()
				accountList := map[uuid.UUID]*account_model.Account{
					merchantID: {UUID: uuid.New(), EODBalance: 100.0},
				}
				
				accountSvc.On("GetMerchantAccounts", mock.Anything, mock.Anything, mock.Anything).Return(accountList, nil)
				accTrxRepo.On("GetBulkAggregateTransactions", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			request: &account_model.GetBulkBalanceRequest{
				MerchantIDs: []uuid.UUID{uuid.New()},
				Usecase:     "test_usecase",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			accountSvcMock := serviceMocks.NewIAccountService(t)
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			
			tc.mocksSetup(accTrxRepoMock, accountSvcMock)

			orchestratorSvc := &OrchestratorService{
				logger:                 loggerMock,
				accountTransactionRepo: accTrxRepoMock,
				accountSvc:             accountSvcMock,
			}

			ctx := context.Background()
			result, err := orchestratorSvc.GetMerchantBulkBalances(ctx, tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if len(tc.request.MerchantIDs) == 0 || tc.name == "SUCCESS: Empty account list" {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
				}
			}

			accTrxRepoMock.AssertExpectations(t)
			accountSvcMock.AssertExpectations(t)
		})
	}
}
