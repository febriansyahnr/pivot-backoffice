package accountService

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCalculateEodBalance(t *testing.T) {
	merchantId := "5e37ed98-7b03-4ad7-a992-6b152aab4fe2"
	ctxTrx := context.WithValue(context.Background(), constant.CtxSyncKey, nil) // Just For Dummy Context

	customerAccount := &account_model.Account{
		UUID:        uuid.New(),
		ReferenceID: util.ParseUUID(merchantId),
		Name:        constant.TypeWallet,
		UserType:    constant.UserTypeCustomer,
		EODBalance:  12500,
		Currency:    "IDR",
	}

	account := []*account_model.Account{
		{
			UUID:        uuid.New(),
			ReferenceID: util.ParseUUID(merchantId),
			Name:        constant.TypePayment,
			EODBalance:  12500,
			Currency:    "IDR",
		},
		customerAccount,
	}

	aggregateResponse := &orchestrator_model.AggregateResponse{
		SumOfDebit:  100000.00,
		SumOfCredit: 100000.00,
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, dailyTrxRepo *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService)
	}{
		{
			name:    "SUCCESS:Calculate EOD Balance",
			wantErr: false,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, dailyTrxRepo *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				customerService.On("FindCustomerByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(&customerModel.GeneralCustomerResponse{MerchantID: uuid.NewString()}, nil)

				trxRepo.On(
					"GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything,
				).Return(aggregateResponse, nil)
				trxRepo.On(
					"GetListOfTransactionReferenceIdsWithPendingStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Return([]string{}, nil)

				logger.On(
					"Info", constant.ValueCtxMockType(), "Balance summary for reference account "+merchantId, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()

				trxRepo.On("GetEarliestUpdatedAt", constant.ValueCtxMockType(), mock.Anything).Return(time.Now().UTC(), nil)
				trxRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctxTrx, nil)
				mockRepo.On("UpdateAccount", constant.ValueCtxMockType(), constant.PtrAccountMockType()).Return(nil)
				trxRepo.On("RearrangeUpdatedAtForTransactionWithPendingStatus", constant.ValueCtxMockType(), mock.Anything, constant.TimeMockType()).Return(nil)
				dailyTrxRepo.On(
					"Upsert", constant.ValueCtxMockType(), mock.AnythingOfType("*dailyAccountTransactionModel.DailyAccountTransaction"),
				).Return(nil)
				trxRepo.On("CommitTransaction", constant.ValueCtxMockType()).Return(nil)
			},
		},
		{
			name:    "SUCCESS:No Transaction Found",
			wantErr: false,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, dailyTrxRepo *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				customerService.On("FindCustomerByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(&customerModel.GeneralCustomerResponse{MerchantID: uuid.NewString()}, nil)
				trxRepo.On(
					"GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything,
				).Return(&orchestrator_model.AggregateResponse{}, nil)

				logger.On(
					"Info", constant.ValueCtxMockType(), "Balance summary for reference account "+merchantId, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()

				trxRepo.On("GetEarliestUpdatedAt", constant.ValueCtxMockType(), mock.Anything).Return(time.Now().UTC(), nil)
				trxRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctxTrx, nil)
				dailyTrxRepo.On(
					"Upsert", constant.ValueCtxMockType(), mock.AnythingOfType("*dailyAccountTransactionModel.DailyAccountTransaction"),
				).Return(nil)
				trxRepo.On("CommitTransaction", constant.ValueCtxMockType()).Return(nil)
			},
		},
		{
			name:    "ERROR:FindAll",
			wantErr: true,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, _ *repositoryMocks.IAccountTransactionRepository, _ *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(nil, constant.ErrSomeErrorForUnitTest)
				logger.On("Error", constant.ValueCtxMockType(), "error when get all accounts", mock.Anything).Return()
			},
		},
		{
			name:    "ERROR: Get Customer",
			wantErr: false,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, _ *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return([]*account_model.Account{customerAccount}, nil)
				customerService.On("FindCustomerByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, constant.ErrSomeErrorForUnitTest)

				logger.On("Error", constant.ValueCtxMockType(), "error when get customer by id", mock.Anything, mock.Anything, mock.Anything).Return()
			},
		},
		{
			name:    "ERROR:GetAggregateTransactions",
			wantErr: true,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, _ *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				trxRepo.On("GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)

				logger.On("Error", constant.ValueCtxMockType(), "error when calculate account balance by merchant, balance and date", mock.Anything).Return()
			},
		},
		{
			name:    "ERROR:GetListOfTransactionReferenceIdsWithPendingStatus",
			wantErr: true,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, _ *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				trxRepo.On(
					"GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything,
				).Return(aggregateResponse, nil)
				logger.On(
					"Info", constant.ValueCtxMockType(), "Balance summary for reference account "+merchantId, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()

				trxRepo.On(
					"GetListOfTransactionReferenceIdsWithPendingStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
				logger.On("Error", constant.ValueCtxMockType(), "error when get list of transaction reference ids with pending status", mock.Anything).Return()
			},
		},
		{
			name:    "ERROR:BeginTransaction",
			wantErr: true,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, _ *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				trxRepo.On(
					"GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything,
				).Return(aggregateResponse, nil)
				logger.On(
					"Info", constant.ValueCtxMockType(), "Balance summary for reference account "+merchantId, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()
				trxRepo.On(
					"GetListOfTransactionReferenceIdsWithPendingStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Return([]string{}, nil)

				trxRepo.On("GetEarliestUpdatedAt", constant.ValueCtxMockType(), mock.Anything).Return(time.Now().UTC(), nil)
				trxRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(nil, constant.ErrSomeErrorForUnitTest)
				logger.On("Error", constant.ValueCtxMockType(), "failed while starting transaction session", mock.Anything).Return()
			},
		},
		{
			name:    "ERROR:UpdateAccount",
			wantErr: true,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, _ *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				trxRepo.On(
					"GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything,
				).Return(aggregateResponse, nil)
				logger.On(
					"Info", constant.ValueCtxMockType(), "Balance summary for reference account "+merchantId, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()
				trxRepo.On(
					"GetListOfTransactionReferenceIdsWithPendingStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Return([]string{}, nil)
				trxRepo.On("GetEarliestUpdatedAt", constant.ValueCtxMockType(), mock.Anything).Return(time.Now().UTC(), nil)
				trxRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctxTrx, nil)

				trxRepo.On("RollbackTransaction", constant.ValueCtxMockType()).Return(constant.ErrSomeErrorForUnitTest)
				mockRepo.On("UpdateAccount", constant.ValueCtxMockType(), constant.PtrAccountMockType()).Return(constant.ErrSomeErrorForUnitTest)
				logger.On("Error", constant.ValueCtxMockType(), "failed to cancel transaction session", mock.Anything).Times(1).Return()
				logger.On("Error", constant.ValueCtxMockType(), "error when update account", mock.Anything).Times(1).Return()
			},
		},
		{
			name:    "ERROR:RearrangeUpdatedAtForTransactionWithPendingStatus",
			wantErr: true,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, _ *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				trxRepo.On(
					"GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything,
				).Return(aggregateResponse, nil)
				logger.On(
					"Info", constant.ValueCtxMockType(), "Balance summary for reference account "+merchantId, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()
				trxRepo.On(
					"GetListOfTransactionReferenceIdsWithPendingStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Return([]string{}, nil)
				trxRepo.On("GetEarliestUpdatedAt", constant.ValueCtxMockType(), mock.Anything).Return(time.Now().UTC(), nil)
				trxRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctxTrx, nil)
				trxRepo.On("RollbackTransaction", constant.ValueCtxMockType()).Return(nil)
				mockRepo.On("UpdateAccount", constant.ValueCtxMockType(), constant.PtrAccountMockType()).Return(nil)

				trxRepo.On("RearrangeUpdatedAtForTransactionWithPendingStatus", constant.ValueCtxMockType(), mock.Anything, constant.TimeMockType()).Return(constant.ErrSomeErrorForUnitTest)
				logger.On("Error", constant.ValueCtxMockType(), "error when rearrange updated at for transaction with pending status", mock.Anything).Return()
			},
		},
		{
			name:    "ERROR:Upsert",
			wantErr: true,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, dailyTrxRepo *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				trxRepo.On(
					"GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything,
				).Return(aggregateResponse, nil)
				logger.On(
					"Info", constant.ValueCtxMockType(), "Balance summary for reference account "+merchantId, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()
				trxRepo.On(
					"GetListOfTransactionReferenceIdsWithPendingStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Return([]string{}, nil)
				trxRepo.On("GetEarliestUpdatedAt", constant.ValueCtxMockType(), mock.Anything).Return(time.Now().UTC(), nil)
				trxRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctxTrx, nil)
				trxRepo.On("RollbackTransaction", constant.ValueCtxMockType()).Return(nil)
				mockRepo.On("UpdateAccount", constant.ValueCtxMockType(), constant.PtrAccountMockType()).Return(nil)
				trxRepo.On("RearrangeUpdatedAtForTransactionWithPendingStatus", constant.ValueCtxMockType(), mock.Anything, constant.TimeMockType()).Return(nil)

				dailyTrxRepo.On(
					"Upsert", constant.ValueCtxMockType(), mock.AnythingOfType("*dailyAccountTransactionModel.DailyAccountTransaction"),
				).Return(constant.ErrSomeErrorForUnitTest)
				logger.On("Error", constant.ValueCtxMockType(), "error when upsert daily account transaction", mock.Anything).Return()
			},
		},
		{
			name:    "ERROR:GetEarliestUpdatedAt",
			wantErr: true,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, _ *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				trxRepo.On(
					"GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything,
				).Return(aggregateResponse, nil)
				logger.On(
					"Info", constant.ValueCtxMockType(), "Balance summary for reference account "+merchantId, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()
				trxRepo.On(
					"GetListOfTransactionReferenceIdsWithPendingStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Return([]string{}, nil)

				trxRepo.On("GetEarliestUpdatedAt", constant.ValueCtxMockType(), mock.Anything).Return(time.Time{}, constant.ErrSomeErrorForUnitTest)
				logger.On("Error", constant.ValueCtxMockType(), "error when get earliest updated at pending transaction", mock.Anything).Return()
			},
		},
		{
			name:    "ERROR:CommitTransaction",
			wantErr: true,
			mockSetup: func(logger *loggerMock.ILogger, mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository, dailyTrxRepo *repositoryMocks.IDailyAccountTransactionRepository, customerService *serviceMock.ICustomerService) {
				mockRepo.On("FindAll", constant.ValueCtxMockType()).Return(account, nil)
				trxRepo.On(
					"GetAggregateTransactions", constant.ValueCtxMockType(), mock.Anything,
				).Return(aggregateResponse, nil)
				logger.On(
					"Info", constant.ValueCtxMockType(), "Balance summary for reference account "+merchantId, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()
				trxRepo.On(
					"GetListOfTransactionReferenceIdsWithPendingStatus", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Return([]string{}, nil)
				trxRepo.On("GetEarliestUpdatedAt", constant.ValueCtxMockType(), mock.Anything).Return(time.Now().UTC(), nil)
				trxRepo.On("BeginTransaction", constant.ValueCtxMockType()).Return(ctxTrx, nil)
				trxRepo.On("RollbackTransaction", constant.ValueCtxMockType()).Return(nil)
				mockRepo.On("UpdateAccount", constant.ValueCtxMockType(), constant.PtrAccountMockType()).Return(nil)
				trxRepo.On("RearrangeUpdatedAtForTransactionWithPendingStatus", constant.ValueCtxMockType(), mock.Anything, constant.TimeMockType()).Return(nil)
				dailyTrxRepo.On(
					"Upsert", constant.ValueCtxMockType(), mock.AnythingOfType("*dailyAccountTransactionModel.DailyAccountTransaction"),
				).Return(nil)

				trxRepo.On("CommitTransaction", constant.ValueCtxMockType()).Return(constant.ErrSomeErrorForUnitTest)
				logger.On("Error", constant.ValueCtxMockType(), "failed while performing transaction", mock.Anything).Return()
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := loggerMock.NewILogger(t)
			accountRepo := repositoryMocks.NewIAccountRepository(t)
			accountTrxRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			dailyTrxRepo := repositoryMocks.NewIDailyAccountTransactionRepository(t)
			customerService := serviceMock.NewICustomerService(t)

			tc.mockSetup(logger, accountRepo, accountTrxRepo, dailyTrxRepo, customerService)

			svc := New(logger, accountTrxRepo, accountRepo, dailyTrxRepo)
			WithCustomerService(svc, customerService)

			if err := svc.CalculateAccountEodBalance(context.Background()); tc.wantErr {
				require.Error(t, err)

			} else {
				assert.NoError(t, err)
			}

			accountRepo.AssertExpectations(t)
			dailyTrxRepo.AssertExpectations(t)
			accountTrxRepo.AssertExpectations(t)
		})
	}
}
