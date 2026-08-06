package accountService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetBalance(t *testing.T) {
	accountResult := &account_model.Account{
		UUID:        uuid.New(),
		ReferenceID: uuid.New(),
		Name:        constant.TypePayment,
		EODBalance:  12500,
		Currency:    "IDR",
	}

	aggregateResponse := &orchestrator_model.AggregateResponse{
		SumOfDebit:  100000.00,
		SumOfCredit: 100000.00,
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository)
	}{
		{
			name:    "SUCCESS: Get Account",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On(
					"GetByUUID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
				).Return(accountResult, nil)

				trxRepo.On(
					"GetAggregateTransactions",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.GetAggregateRequest"),
				).Return(aggregateResponse, nil)
			},
		},
		{
			name:    "ERROR: GetByUUID",
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On(
					"GetByUUID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)

			},
		},
		{
			name:    "ERROR: GetAggregateTransactions",
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IAccountRepository, trxRepo *repositoryMocks.IAccountTransactionRepository) {
				mockRepo.On(
					"GetByUUID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
				).Return(accountResult, nil)

				trxRepo.On(
					"GetAggregateTransactions",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*orchestrator_model.GetAggregateRequest"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)

			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIAccountRepository(t)
			trxRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLog, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			ctx := context.Background()
			tc.mockSetup(mockRepo, trxRepo)

			svc := New(mockLog, trxRepo, mockRepo, nil)
			response, err := svc.GetAccount(ctx, uuid.New())
			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, response)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			mockRepo.AssertExpectations(t)
			trxRepo.AssertExpectations(t)

		})
	}
}
