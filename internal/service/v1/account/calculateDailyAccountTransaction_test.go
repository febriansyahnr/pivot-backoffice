package accountService

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	dailyAccountTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/dailyAccountTransaction"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestCalculateDailyAccountTransaction(t *testing.T) {
	location, _ := time.LoadLocation("Asia/Jakarta")

	account := &accountModel.Account{
		UUID:        uuid.New(),
		ReferenceID: uuid.New(),
		CreatedAt:   time.Now().AddDate(-1, 0, 0),
		EODBalance:  0,
	}

	customerAccount := &accountModel.Account{
		UUID:        uuid.New(),
		ReferenceID: uuid.New(),
		Name:        constant.TypeWallet,
		UserType:    constant.UserTypeCustomer,
		EODBalance:  12500,
		Currency:    "IDR",
	}

	latestDailyAccountTransaction := &dailyAccountTransactionModel.DailyAccountTransaction{
		EODBalance: 10000,
		Date:       time.Now().AddDate(0, 0, -1),
		Timezone:   "Asia/Jakarta",
	}

	aggregateResponse := &orchestratorModel.AggregateResponse{
		SumOfDebit:    5000,
		SumOfCredit:   10000,
		CountOfDebit:  5,
		CountOfCredit: 3,
	}

	testCases := []struct {
		name      string
		mockSetup func(
			mockAccountRepo *repositoryMocks.IAccountRepository,
			mockDailyAccountRepo *repositoryMocks.IDailyAccountTransactionRepository,
			mockTransactionRepo *repositoryMocks.IAccountTransactionRepository,
			customerService *serviceMock.ICustomerService,
		)
		wantErr bool
	}{
		{
			name:    "SUCCESS: Calculate daily account transaction",
			wantErr: false,
			mockSetup: func(
				mockAccountRepo *repositoryMocks.IAccountRepository,
				mockDailyAccountRepo *repositoryMocks.IDailyAccountTransactionRepository,
				mockTransactionRepo *repositoryMocks.IAccountTransactionRepository,
				customerService *serviceMock.ICustomerService,
			) {
				mockAccountRepo.On("FindAll", mock.Anything).
					Return([]*accountModel.Account{account, customerAccount}, nil)

				customerService.On("FindCustomerByID", mock.Anything, mock.Anything).Return(&customerModel.GeneralCustomerResponse{MerchantID: uuid.NewString()}, nil)

				mockDailyAccountRepo.On("FindLatestByAccountIDAndTimezone", mock.Anything, mock.Anything, mock.Anything).
					Return(latestDailyAccountTransaction, nil)

				mockTransactionRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).
					Return(aggregateResponse, nil)

				mockDailyAccountRepo.On("Upsert", mock.Anything, mock.Anything).
					Return(nil)
			},
		},
		{
			name:    "ERROR: FindAll",
			wantErr: true,
			mockSetup: func(
				mockAccountRepo *repositoryMocks.IAccountRepository,
				mockDailyAccountRepo *repositoryMocks.IDailyAccountTransactionRepository,
				mockTransactionRepo *repositoryMocks.IAccountTransactionRepository,
				customerService *serviceMock.ICustomerService,
			) {
				mockAccountRepo.On("FindAll", mock.Anything).
					Return(nil, assert.AnError)
			},
		},
		{
			name:    "ERROR: FindLatestByAccountIDAndTimezone",
			wantErr: true,
			mockSetup: func(
				mockAccountRepo *repositoryMocks.IAccountRepository,
				mockDailyAccountRepo *repositoryMocks.IDailyAccountTransactionRepository,
				mockTransactionRepo *repositoryMocks.IAccountTransactionRepository,
				customerService *serviceMock.ICustomerService,
			) {
				mockAccountRepo.On("FindAll", mock.Anything).
					Return([]*accountModel.Account{account}, nil)

				mockDailyAccountRepo.On("FindLatestByAccountIDAndTimezone", mock.Anything, account.UUID.String(), location.String()).
					Return(nil, assert.AnError)
			},
		},
		{
			name:    "ERROR: Get Customer",
			wantErr: false,
			mockSetup: func(
				mockAccountRepo *repositoryMocks.IAccountRepository,
				mockDailyAccountRepo *repositoryMocks.IDailyAccountTransactionRepository,
				mockTransactionRepo *repositoryMocks.IAccountTransactionRepository,
				customerService *serviceMock.ICustomerService,
			) {
				mockAccountRepo.On("FindAll", mock.Anything).
					Return([]*accountModel.Account{customerAccount}, nil)

				mockDailyAccountRepo.On("FindLatestByAccountIDAndTimezone", mock.Anything, mock.Anything, mock.Anything).
					Return(latestDailyAccountTransaction, nil)

				customerService.On("FindCustomerByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
		},
		{
			name:    "ERROR: GetAggregateTransactions",
			wantErr: true,
			mockSetup: func(
				mockAccountRepo *repositoryMocks.IAccountRepository,
				mockDailyAccountRepo *repositoryMocks.IDailyAccountTransactionRepository,
				mockTransactionRepo *repositoryMocks.IAccountTransactionRepository,
				customerService *serviceMock.ICustomerService,
			) {
				mockAccountRepo.On("FindAll", mock.Anything).
					Return([]*accountModel.Account{account}, nil)

				mockDailyAccountRepo.On("FindLatestByAccountIDAndTimezone", mock.Anything, account.UUID.String(), location.String()).
					Return(latestDailyAccountTransaction, nil)

				mockTransactionRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
		},
		{
			name:    "ERROR: Upsert",
			wantErr: true,
			mockSetup: func(
				mockAccountRepo *repositoryMocks.IAccountRepository,
				mockDailyAccountRepo *repositoryMocks.IDailyAccountTransactionRepository,
				mockTransactionRepo *repositoryMocks.IAccountTransactionRepository,
				customerService *serviceMock.ICustomerService,
			) {
				mockAccountRepo.On("FindAll", mock.Anything).
					Return([]*accountModel.Account{account}, nil)

				mockDailyAccountRepo.On("FindLatestByAccountIDAndTimezone", mock.Anything, account.UUID.String(), location.String()).
					Return(latestDailyAccountTransaction, nil)

				mockTransactionRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).
					Return(aggregateResponse, nil)

				mockDailyAccountRepo.On("Upsert", mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAccountRepo := repositoryMocks.NewIAccountRepository(t)
			mockDailyAccountRepo := repositoryMocks.NewIDailyAccountTransactionRepository(t)
			mockTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCustomerService := serviceMock.NewICustomerService(t)

			svc := New(mockLogger, mockTransactionRepo, mockAccountRepo, mockDailyAccountRepo)
			WithCustomerService(svc, mockCustomerService)

			tc.mockSetup(mockAccountRepo, mockDailyAccountRepo, mockTransactionRepo, mockCustomerService)

			err := svc.CalculateDailyAccountTransaction(context.Background(), location)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockAccountRepo.AssertExpectations(t)
			mockDailyAccountRepo.AssertExpectations(t)
			mockTransactionRepo.AssertExpectations(t)
		})
	}
}
