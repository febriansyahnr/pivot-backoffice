package orchestrator_service

import (
	"context"
	"errors"
	"testing"
	"time"

	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"

	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAccountTransaction(t *testing.T) {
	merchantID := uuid.New()
	createdAccountTransaction := &orchestrator_model.CreateAccountTransactionRequest{
		ReferenceID:          "TRXVA1234567890",
		MerchantID:           merchantID,
		Currency:             "IDR",
		Credit:               12500,
		Debit:                0,
		Type:                 constant.TypePayment,
		Channel:              constant.ChannelVirtualAccount,
		Status:               constant.StatusSuccess,
		Remarks:              "Transaction VA - TRXVA1234567890",
		TransactionTimestamp: time.Now(),
	}

	createdAccountTransactionForTopUp := &orchestrator_model.CreateAccountTransactionRequest{
		ReferenceID:          uuid.NewString(),
		MerchantID:           merchantID,
		Currency:             "IDR",
		Credit:               12500,
		Debit:                0,
		Type:                 constant.TypeTopUp,
		Channel:              constant.ChannelVirtualAccount,
		Status:               constant.StatusSuccess,
		Remarks:              "Transaction VA - TRXVA1234567890",
		TransactionTimestamp: time.Now(),
	}

	merchantAccount := &account_model.Account{
		UUID:        uuid.New(),
		ReferenceID: merchantID,
		Name:        "Payment - Virtual Account",
		EODBalance:  12500,
		Currency:    "IDR",

		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	testCases := []struct {
		name       string
		input      *orchestrator_model.CreateAccountTransactionRequest
		mocksSetup func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
			accountRepo *repositoryMocks.IAccountRepository)
		wantErr bool
	}{
		{
			name:  "SUCCESS: successfully create transaction with create balance",
			input: createdAccountTransaction,
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				accountRepo *repositoryMocks.IAccountRepository) {
				createdAccountTransaction.Type = constant.TypeTopUp
				accTrxRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				accountRepo.On("FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(), mock.AnythingOfType("string"),
				).Return(nil, nil)

				accountRepo.On("Create", mock.Anything, constant.PtrAccountMockType()).
					Return(nil)

				accTrxRepo.On("Create", mock.Anything, mock.AnythingOfType("*orchestrator_model.AccountTransaction")).
					Return(nil)

				accTrxRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "SUCCESS: successfully create transaction without create balance",
			input: createdAccountTransaction,
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				accountRepo *repositoryMocks.IAccountRepository) {

				createdAccountTransaction.Type = constant.TypePayment
				accTrxRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				accountRepo.On("FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(), mock.AnythingOfType("string"),
				).Return(merchantAccount, nil)

				accTrxRepo.On("Create", mock.Anything, mock.AnythingOfType("*orchestrator_model.AccountTransaction")).
					Return(nil)

				accTrxRepo.On("CommitTransaction", mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: Begin Transaction",
			input: createdAccountTransaction,
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				accountRepo *repositoryMocks.IAccountRepository) {
				accTrxRepo.On("BeginTransaction", mock.Anything).
					Return(nil, errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name:  "ERROR: FindMerchantAccountByName",
			input: createdAccountTransaction,
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				accountRepo *repositoryMocks.IAccountRepository) {
				accTrxRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				accountRepo.On("FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(), mock.AnythingOfType("string"),
				).Return(nil, errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Create Balance",
			input: createdAccountTransaction,
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				accountRepo *repositoryMocks.IAccountRepository) {
				accTrxRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				accountRepo.On("FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(), mock.AnythingOfType("string"),
				).Return(nil, nil)

				accountRepo.On("Create", mock.Anything, constant.PtrAccountMockType()).
					Return(errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Create Account Transaction",
			input: createdAccountTransaction,
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				accountRepo *repositoryMocks.IAccountRepository) {
				accTrxRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				accountRepo.On("FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(), mock.AnythingOfType("string"),
				).Return(merchantAccount, nil)

				accTrxRepo.On("Create", mock.Anything, mock.AnythingOfType("*orchestrator_model.AccountTransaction")).
					Return(errors.New("some-error"))

			},
			wantErr: true,
		},
		{
			name:  "ERROR: Create Account Transaction For Disbursement Top Up",
			input: createdAccountTransactionForTopUp,
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				accountRepo *repositoryMocks.IAccountRepository) {
				accTrxRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				accountRepo.On("FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(), mock.AnythingOfType("string"),
				).Return(merchantAccount, nil)

				accTrxRepo.On("Create", mock.Anything, mock.AnythingOfType("*orchestrator_model.AccountTransaction")).
					Return(errors.New("some-error"))

			},
			wantErr: true,
		},
		{
			name:  "ERROR: Commit Transaction",
			input: createdAccountTransaction,
			mocksSetup: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				accountRepo *repositoryMocks.IAccountRepository) {
				accTrxRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)

				accountRepo.On("FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference), constant.UuidMockType(), mock.AnythingOfType("string"),
				).Return(merchantAccount, nil)

				accTrxRepo.On("Create", mock.Anything, mock.AnythingOfType("*orchestrator_model.AccountTransaction")).
					Return(nil)

				accTrxRepo.On("CommitTransaction", mock.Anything).
					Return(errors.New("some-error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			accountRepoMock := repositoryMocks.NewIAccountRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mocksSetup(accTrxRepoMock, accountRepoMock)

			accTrxSvc := New(mockLogger, nil, accTrxRepoMock, accountRepoMock)
			ctx := context.Background()
			err := accTrxSvc.CreateAccountTransaction(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			accTrxRepoMock.AssertExpectations(t)
		})
	}
}
