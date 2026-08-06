package ledgerService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLedgerBalance(t *testing.T) {

	testCases := []struct {
		Name      string
		MockSetup func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository)
		WantErr   bool
	}{
		{
			Name: "SUCCESS: Get Ledger Balance",
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository) {
				mockAccRepo.On("GetByUUID",
					mock.Anything,
					mock.Anything,
				).Return(&account_model.Account{
					EODBalance: 100,
				}, nil)
				mockRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).Return(&orchestrator_model.AggregateResponse{
					SumOfCredit: 100,
					SumOfDebit:  100,
				}, nil)
			},
			WantErr: false,
		},
		{
			Name: "Error: Get account",
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository) {
				mockAccRepo.On("GetByUUID",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))
			},
			WantErr: true,
		},
		{
			Name: "Error: Account not found",
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository) {
				mockAccRepo.On("GetByUUID",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
			},
			WantErr: true,
		},
		{
			Name: "Error: Get aggregate transactions",
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository) {
				mockAccRepo.On("GetByUUID",
					mock.Anything,
					mock.Anything,
				).Return(&account_model.Account{
					EODBalance: 100,
				}, nil)
				mockRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			accTrxRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			accRepo := repositoryMocks.NewIAccountRepository(t)

			tc.MockSetup(accTrxRepo, accRepo)
			svc := New(loggerMock, accTrxRepo, accRepo, nil, nil, nil)
			balancer, err := svc.GetLedgerBalance(context.Background(), uuid.New())

			if tc.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, balancer)
			}
		})

	}
}

func TestCalculateBulkLedgerBalance(t *testing.T) {
	testCases := []struct {
		Name            string
		Request         *account_model.CalculateBulkLedgerBalanceRequest
		MockSetup       func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository)
		WantErr         bool
		ExpectedBalance float64
	}{
		{
			Name: "SUCCESS: Calculate bulk balance",
			Request: &account_model.CalculateBulkLedgerBalanceRequest{
				MerchantID: uuid.New().String(),
				AccountIDs: []string{uuid.New().String(), uuid.New().String()},
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository) {
				mockAccRepo.On("GetByIDs", mock.Anything, mock.Anything).Return([]*account_model.Account{
					{EODBalance: 100},
					{EODBalance: 200},
				}, nil)
				mockRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).Return(&orchestrator_model.AggregateResponse{
					SumOfCredit: 50,
					SumOfDebit:  30,
				}, nil)
			},
			WantErr:         false,
			ExpectedBalance: 320, // 100 + 200 + 50 - 30
		},
		{
			Name: "SUCCESS: Empty account list returns zero balance",
			Request: &account_model.CalculateBulkLedgerBalanceRequest{
				MerchantID: uuid.New().String(),
				AccountIDs: []string{},
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository) {
				mockAccRepo.On("GetByIDs", mock.Anything, mock.Anything).Return([]*account_model.Account{}, nil)
			},
			WantErr:         false,
			ExpectedBalance: 0,
		},
		{
			Name: "Error: Get accounts by IDs fails",
			Request: &account_model.CalculateBulkLedgerBalanceRequest{
				MerchantID: uuid.New().String(),
				AccountIDs: []string{uuid.New().String()},
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository) {
				mockAccRepo.On("GetByIDs", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			WantErr: true,
		},
		{
			Name: "Error: Get aggregate transactions fails",
			Request: &account_model.CalculateBulkLedgerBalanceRequest{
				MerchantID: uuid.New().String(),
				AccountIDs: []string{uuid.New().String()},
			},
			MockSetup: func(mockRepo *repositoryMocks.IAccountTransactionRepository, mockAccRepo *repositoryMocks.IAccountRepository) {
				mockAccRepo.On("GetByIDs", mock.Anything, mock.Anything).Return([]*account_model.Account{
					{EODBalance: 100},
				}, nil)
				mockRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).Return(nil, errors.New("aggregate error"))
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			accTrxRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			accRepo := repositoryMocks.NewIAccountRepository(t)

			tc.MockSetup(accTrxRepo, accRepo)
			svc := New(loggerMock, accTrxRepo, accRepo, nil, nil, nil)
			balance, err := svc.CalculateBulkLedgerBalance(context.Background(), tc.Request)

			if tc.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, balance)
				assert.Equal(t, tc.ExpectedBalance, balance.Balance)
			}
		})
	}
}
