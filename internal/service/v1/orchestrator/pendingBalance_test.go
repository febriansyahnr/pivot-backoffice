package orchestrator_service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPendingBalance(t *testing.T) {
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
					mock.Anything,
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(&account_model.Account{}, nil)

				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(time.Now(), nil)

				accTrxRepo.On(
					"GetAggregateTransactions",
					mock.Anything,
					mock.Anything,
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
					mock.Anything,
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(nil, nil)

				balanceRepo.On(
					"Create",
					mock.Anything,
					constant.PtrAccountMockType(),
				).Return(nil)

				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(time.Now(), nil)

				accTrxRepo.On(
					"GetAggregateTransactions",
					mock.Anything,
					mock.Anything,
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
					mock.Anything,
					constant.UuidMockType(),
					constant.StringMockType(),
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
					mock.Anything,
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(&account_model.Account{}, nil)

				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(time.Now(), nil)

				accTrxRepo.On(
					"GetAggregateTransactions",
					mock.Anything,
					mock.Anything,
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
					mock.Anything,
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(nil, nil)

				balanceRepo.On(
					"Create",
					mock.Anything,
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
			_, err := accTrxSvc.GetPendingBalance(ctx, tc.inputMerchantID, constant.TypeDisbursement)
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

func TestGetPendingBalanceComprehensive(t *testing.T) {
	merchantID := uuid.New()
	accountID := uuid.New()
	mockTime := time.Now()

	testCases := []struct {
		name        string
		balanceName string
		mocksSetup  func(
			accTrxRepo *repositoryMocks.IAccountTransactionRepository,
			balanceRepo *repositoryMocks.IAccountRepository,
			redisRepo *redisExtMocks.IRedisExt,
		)
		expectedBalance float64
		wantErr         bool
	}{
		{
			name:        "SUCCESS: Payment type - CalculatePendingBalance",
			balanceName: constant.ProductPayment,
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.Anything,
					merchantID,
					constant.ProductPayment,
				).Return(&account_model.Account{UUID: accountID}, nil)

				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(mockTime, nil)

				accTrxRepo.On(
					"CalculatePendingBalance",
					mock.Anything,
					mock.Anything,
				).Return(1500.0, nil)
			},
			expectedBalance: 1500.0,
			wantErr:         false,
		},
		{
			name:        "SUCCESS: Wallet type - CalculatePendingBalance",
			balanceName: constant.ReferenceWallet,
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.Anything,
					merchantID,
					constant.ReferenceWallet,
				).Return(&account_model.Account{UUID: accountID}, nil)

				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(mockTime, nil)

				accTrxRepo.On(
					"CalculatePendingBalance",
					mock.Anything,
					mock.Anything,
				).Return(2500.0, nil)
			},
			expectedBalance: 2500.0,
			wantErr:         false,
		},
		{
			name:        "SUCCESS: Regular balance type - GetAggregateTransactions",
			balanceName: constant.TypeDisbursement,
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.Anything,
					merchantID,
					constant.TypeDisbursement,
				).Return(&account_model.Account{UUID: accountID}, nil)

				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(mockTime, nil)

				accTrxRepo.On(
					"GetAggregateTransactions",
					mock.Anything,
					mock.Anything,
				).Return(&orchestratorModel.AggregateResponse{
					SumOfCredit: 5000.0,
					SumOfDebit:  2000.0,
				}, nil)
			},
			expectedBalance: 3000.0,
			wantErr:         false,
		},
		{
			name:        "ERROR: GetCachedEarliestUpdatedAt error",
			balanceName: constant.TypeDisbursement,
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.Anything,
					merchantID,
					constant.TypeDisbursement,
				).Return(&account_model.Account{UUID: accountID}, nil)

				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(time.Time{}, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:        "ERROR: CalculatePendingBalance error for payment type",
			balanceName: constant.ProductPayment,
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.Anything,
					merchantID,
					constant.ProductPayment,
				).Return(&account_model.Account{UUID: accountID}, nil)

				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(mockTime, nil)

				accTrxRepo.On(
					"CalculatePendingBalance",
					mock.Anything,
					mock.Anything,
				).Return(0.0, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			balanceRepoMock := repositoryMocks.NewIAccountRepository(t)
			redisRepoMock := redisExtMocks.NewIRedisExt(t)
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mocksSetup(accTrxRepoMock, balanceRepoMock, redisRepoMock)

			service := New(loggerMock, nil, accTrxRepoMock, balanceRepoMock)
			ctx := context.Background()

			balance, err := service.GetPendingBalance(ctx, merchantID.String(), tc.balanceName)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBalance, balance)
			}

			accTrxRepoMock.AssertExpectations(t)
			balanceRepoMock.AssertExpectations(t)
		})
	}
}

func TestGetCachedEarliestUpdatedAt(t *testing.T) {
	merchantID := uuid.New()
	accountID := uuid.New()
	mockTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	request := &orchestratorModel.GetAggregateRequest{
		MerchantID: merchantID,
		AccountID:  accountID,
		Statuses:   []string{constant.StatusPending},
	}

	testCases := []struct {
		name       string
		mocksSetup func(
			accTrxRepo *repositoryMocks.IAccountTransactionRepository,
			balanceRepo *repositoryMocks.IAccountRepository,
			redisRepo *redisExtMocks.IRedisExt,
		)
		setupService func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, balanceRepo *repositoryMocks.IAccountRepository, redisRepo *redisExtMocks.IRedisExt) *OrchestratorService
		expected     time.Time
		wantErr      bool
	}{
		{
			name: "SUCCESS: Redis not configured - direct DB call",
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(mockTime, nil)
			},
			setupService: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, balanceRepo *repositoryMocks.IAccountRepository, redisRepo *redisExtMocks.IRedisExt) *OrchestratorService {
				loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
				return &OrchestratorService{
					logger:                 loggerMock,
					accountTransactionRepo: accTrxRepo,
					accountRepo:            balanceRepo,
					redis:                  nil,
				}
			},
			expected: mockTime,
			wantErr:  false,
		},
		{
			name: "SUCCESS: Cache hit",
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				cachedTimeBytes, _ := json.Marshal(mockTime)
				redisRepoMock := &redis.StringCmd{}
				redisRepoMock.SetVal(string(cachedTimeBytes))

				redisRepo.On(
					"Get",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(redisRepoMock)
			},
			setupService: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, balanceRepo *repositoryMocks.IAccountRepository, redisRepo *redisExtMocks.IRedisExt) *OrchestratorService {
				loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
				return &OrchestratorService{
					logger:                 loggerMock,
					accountTransactionRepo: accTrxRepo,
					accountRepo:            balanceRepo,
					redis:                  redisRepo,
				}
			},
			expected: mockTime,
			wantErr:  false,
		},
		{
			name: "SUCCESS: Cache miss - fetch from DB and set cache",
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				// Redis Get returns Nil (cache miss)
				redisRepoMock := &redis.StringCmd{}
				redisRepoMock.SetErr(redis.Nil)
				redisRepo.On(
					"Get",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(redisRepoMock)

				// DB call to get earliest updated at
				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(mockTime, nil)

				// Redis Set to cache the result
				redisSetMock := &redis.StatusCmd{}
				redisRepo.On(
					"Set",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					24*time.Hour,
				).Return(redisSetMock)
			},
			setupService: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, balanceRepo *repositoryMocks.IAccountRepository, redisRepo *redisExtMocks.IRedisExt) *OrchestratorService {
				loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
				return &OrchestratorService{
					logger:                 loggerMock,
					accountTransactionRepo: accTrxRepo,
					accountRepo:            balanceRepo,
					redis:                  redisRepo,
				}
			},
			expected: mockTime,
			wantErr:  false,
		},
		{
			name: "SUCCESS: Cache get error - fallback to DB",
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				// Redis Get returns error
				redisRepoMock := &redis.StringCmd{}
				redisRepoMock.SetErr(errors.New("redis connection error"))
				redisRepo.On(
					"Get",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(redisRepoMock)

				// DB call to get earliest updated at
				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(mockTime, nil)

				// Redis Set to cache the result
				redisSetMock := &redis.StatusCmd{}
				redisRepo.On(
					"Set",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					24*time.Hour,
				).Return(redisSetMock)
			},
			setupService: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, balanceRepo *repositoryMocks.IAccountRepository, redisRepo *redisExtMocks.IRedisExt) *OrchestratorService {
				loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
				return &OrchestratorService{
					logger:                 loggerMock,
					accountTransactionRepo: accTrxRepo,
					accountRepo:            balanceRepo,
					redis:                  redisRepo,
				}
			},
			expected: mockTime,
			wantErr:  false,
		},
		{
			name: "SUCCESS: Invalid cached data - fallback to DB",
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				// Redis Get returns invalid JSON
				redisRepoMock := &redis.StringCmd{}
				redisRepoMock.SetVal("invalid json data")
				redisRepo.On(
					"Get",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(redisRepoMock)

				// DB call to get earliest updated at
				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(mockTime, nil)

				// Redis Set to cache the result
				redisSetMock := &redis.StatusCmd{}
				redisRepo.On(
					"Set",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					24*time.Hour,
				).Return(redisSetMock)
			},
			setupService: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, balanceRepo *repositoryMocks.IAccountRepository, redisRepo *redisExtMocks.IRedisExt) *OrchestratorService {
				loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
				return &OrchestratorService{
					logger:                 loggerMock,
					accountTransactionRepo: accTrxRepo,
					accountRepo:            balanceRepo,
					redis:                  redisRepo,
				}
			},
			expected: mockTime,
			wantErr:  false,
		},
		{
			name: "ERROR: DB error",
			mocksSetup: func(
				accTrxRepo *repositoryMocks.IAccountTransactionRepository,
				balanceRepo *repositoryMocks.IAccountRepository,
				redisRepo *redisExtMocks.IRedisExt,
			) {
				// Redis Get returns Nil (cache miss)
				redisRepoMock := &redis.StringCmd{}
				redisRepoMock.SetErr(redis.Nil)
				redisRepo.On(
					"Get",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(redisRepoMock)

				// DB call returns error
				accTrxRepo.On(
					"GetEarliestUpdatedAt",
					mock.Anything,
					mock.Anything,
				).Return(time.Time{}, constant.ErrSomeErrorForUnitTest)
			},
			setupService: func(accTrxRepo *repositoryMocks.IAccountTransactionRepository, balanceRepo *repositoryMocks.IAccountRepository, redisRepo *redisExtMocks.IRedisExt) *OrchestratorService {
				loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
				return &OrchestratorService{
					logger:                 loggerMock,
					accountTransactionRepo: accTrxRepo,
					accountRepo:            balanceRepo,
					redis:                  redisRepo,
				}
			},
			expected: time.Time{},
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			accTrxRepoMock := repositoryMocks.NewIAccountTransactionRepository(t)
			balanceRepoMock := repositoryMocks.NewIAccountRepository(t)
			redisRepoMock := redisExtMocks.NewIRedisExt(t)

			tc.mocksSetup(accTrxRepoMock, balanceRepoMock, redisRepoMock)

			service := tc.setupService(accTrxRepoMock, balanceRepoMock, redisRepoMock)
			ctx := context.Background()

			result, err := service.GetCachedEarliestUpdatedAt(ctx, request)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			accTrxRepoMock.AssertExpectations(t)
			balanceRepoMock.AssertExpectations(t)
			if service.redis != nil {
				redisRepoMock.AssertExpectations(t)
			}
		})
	}
}

func TestGenerateRequestHash(t *testing.T) {
	merchantID := uuid.New()
	accountID := uuid.New()

	loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	service := &OrchestratorService{
		logger: loggerMock,
	}

	testCases := []struct {
		name     string
		request  *orchestratorModel.GetAggregateRequest
		wantErr  bool
		validate func(t *testing.T, hash string)
	}{
		{
			name: "SUCCESS: Generate hash for basic request",
			request: &orchestratorModel.GetAggregateRequest{
				MerchantID:                  merchantID,
				AccountID:                   accountID,
				Statuses:                    []string{constant.StatusPending},
				IncludeFeeIndirectDeduction: false,
			},
			wantErr: false,
			validate: func(t *testing.T, hash string) {
				assert.NotEmpty(t, hash)
				assert.Len(t, hash, 16)
			},
		},
		{
			name: "SUCCESS: Generate hash with account IDs",
			request: &orchestratorModel.GetAggregateRequest{
				MerchantID:                  merchantID,
				AccountID:                   accountID,
				AccountIDs:                  []string{uuid.NewString(), uuid.NewString()},
				Statuses:                    []string{constant.StatusPending, constant.StatusSuccess},
				IncludeFeeIndirectDeduction: true,
			},
			wantErr: false,
			validate: func(t *testing.T, hash string) {
				assert.NotEmpty(t, hash)
				assert.Len(t, hash, 16)
			},
		},
		{
			name: "SUCCESS: Same request produces same hash",
			request: &orchestratorModel.GetAggregateRequest{
				MerchantID:                  merchantID,
				AccountID:                   accountID,
				Statuses:                    []string{constant.StatusPending},
				IncludeFeeIndirectDeduction: false,
			},
			wantErr: false,
			validate: func(t *testing.T, hash string) {
				// Generate the same hash again
				hash2, err := service.generateRequestHash(&orchestratorModel.GetAggregateRequest{
					MerchantID:                  merchantID,
					AccountID:                   accountID,
					Statuses:                    []string{constant.StatusPending},
					IncludeFeeIndirectDeduction: false,
				})
				assert.NoError(t, err)
				assert.Equal(t, hash, hash2)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := service.generateRequestHash(tc.request)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tc.validate(t, hash)
			}
		})
	}
}
