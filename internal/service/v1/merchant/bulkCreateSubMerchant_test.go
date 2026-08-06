package merchant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRabbitMQ "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockVault "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/go-redis/redismock/v9"
	"github.com/jmoiron/sqlx"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBulkCreateSubMerchant(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		request        *merchantModel.BulkCreateSubMerchantRequest
		setupMocks     func(repo *mockRepo.IMerchantRepository, redisClientMock *mockRedis.IRedisExt)
		expectedError  bool
		expectedResult *merchantModel.BulkCreateSubMerchantResponse
	}{
		{
			name: "SUCCESS: Valid bulk creation request",
			request: &merchantModel.BulkCreateSubMerchantRequest{
				MerchantId: "parent-merchant-123",
				KYCType:    constant.MerchantKYCTypeKYC,
				FileName:   "test.csv",
				SubmerchantDetails: [][]string{
					{
						"Test Merchant", "TM", "logo.png", "test@merchant.com", "08123456789",
						"ID", "INDIVIDUAL", "PT", "John Doe", "08987654321", "john@test.com",
						"Jl Test 123", "12345", "1234567890", "BCA",
					},
				},
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, redisClientMock *mockRedis.IRedisExt) {
				repo.On("FindMerchantByID", mock.Anything, "parent-merchant-123").Return(&merchantModel.Merchant{
					UUID: "parent-merchant-123",
					Name: "Parent Merchant",
				}, nil)
			},
			expectedError: false,
			expectedResult: &merchantModel.BulkCreateSubMerchantResponse{
				TotalFailed: 0,
			},
		},
		{
			name: "SUCCESS: Empty submerchant details",
			request: &merchantModel.BulkCreateSubMerchantRequest{
				MerchantId:         "parent-merchant-123",
				KYCType:            constant.MerchantKYCTypeNonKYC,
				FileName:           "empty.csv",
				SubmerchantDetails: [][]string{},
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, redisClientMock *mockRedis.IRedisExt) {
				repo.On("FindMerchantByID", mock.Anything, "parent-merchant-123").Return(&merchantModel.Merchant{
					UUID: "parent-merchant-123",
					Name: "Parent Merchant",
				}, nil)
			},
			expectedError: false,
			expectedResult: &merchantModel.BulkCreateSubMerchantResponse{
				TotalFailed: 0,
				Results:     []merchantModel.BulkCreateSubMerchantDetailResponse{},
			},
		},
		{
			name: "ERROR: Parent merchant not found",
			request: &merchantModel.BulkCreateSubMerchantRequest{
				MerchantId: "non-existent-merchant",
				KYCType:    constant.MerchantKYCTypeKYC,
				FileName:   "test.csv",
				SubmerchantDetails: [][]string{
					{"Test", "T", "", "test@test.com", "08123", "ID", "IND", "PT", "PIC", "08456", "pic@test.com", "Addr", "12345", "1234567890", "BCA"},
				},
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, redisClientMock *mockRedis.IRedisExt) {
				repo.On("FindMerchantByID", mock.Anything, "non-existent-merchant").Return(nil, nil)
			},
			expectedError: true,
		},
		{
			name: "ERROR: Database error when finding parent merchant",
			request: &merchantModel.BulkCreateSubMerchantRequest{
				MerchantId: "merchant-with-db-error",
				KYCType:    constant.MerchantKYCTypeKYC,
				FileName:   "test.csv",
				SubmerchantDetails: [][]string{
					{"Test", "T", "", "test@test.com", "08123", "ID", "IND", "PT", "PIC", "08456", "pic@test.com", "Addr", "12345", "1234567890", "BCA"},
				},
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, redisClientMock *mockRedis.IRedisExt) {
				repo.On("FindMerchantByID", mock.Anything, "merchant-with-db-error").Return(nil, errors.New("database connection error"))
			},
			expectedError: true,
		},
		{
			name: "MIXED: Some validation failures",
			request: &merchantModel.BulkCreateSubMerchantRequest{
				MerchantId: "parent-merchant-123",
				KYCType:    constant.MerchantKYCTypeKYC,
				FileName:   "test.csv",
				SubmerchantDetails: [][]string{
					// Valid submerchant
					{
						"Valid Merchant", "VM", "logo.png", "valid@test.com", "08111111111",
						"ID", "INDIVIDUAL", "PT", "Valid PIC", "08222222222", "valid.pic@test.com",
						"Valid Address", "11111", "1111111111", "BCA",
					},
					// Invalid submerchant - missing required fields
					{
						"", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
					},
				},
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, redisClientMock *mockRedis.IRedisExt) {
				repo.On("FindMerchantByID", mock.Anything, "parent-merchant-123").Return(&merchantModel.Merchant{
					UUID: "parent-merchant-123",
					Name: "Parent Merchant",
				}, nil)

				// redisClientMock.ExpectZAdd()
			},
			expectedError: false,
			expectedResult: &merchantModel.BulkCreateSubMerchantResponse{
				TotalFailed: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			repo := mockRepo.NewIMerchantRepository(t)
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			userRepo := mockRepo.NewIUserRepository(t)
			jwtSvc := mockJWT.NewIJwt(t)
			encryptMock := mockEncrypt.NewICrypto(t)
			rabbitMqExt := mockRabbitMQ.NewRabbitMQExt(t)
			validator := validatorExt.New()
			redisClientMock := mockRedis.NewIRedisExt(t)
			// Setup mocks
			tt.setupMocks(repo, redisClientMock)

			// Mock RabbitMQ publish if not expecting error
			if !tt.expectedError {
				rabbitMqExt.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil).Maybe()
			}

			// Create service instance
			s := New(
				repo, logger, userRepo, jwtSvc, rabbitMqExt, encryptMock,
				WithRedisClient(redisClientMock),
				WithValidator(validator),
			)

			// Execute the function
			result, err := s.BulkCreateSubMerchant(ctx, tt.request)

			// Assertions
			if tt.expectedError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)

				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult.TotalFailed, result.TotalFailed)
					assert.NotEmpty(t, result.ID) // Session ID should be generated
				}

				// Verify that validation failures are captured
				for _, resultDetail := range result.Results {
					if resultDetail.Error != "" {
						assert.NotEmpty(t, resultDetail.Error)
						assert.Empty(t, resultDetail.MerchantID)
					}
				}
			}

			// Assert mock expectations
			repo.AssertExpectations(t)
			rabbitMqExt.AssertExpectations(t)
		})
	}
}

func TestBulkCreateSubMerchantBatchProcessing(t *testing.T) {
	ctx := context.Background()

	// Test case for verifying batch processing with more than batch size (10) records
	batchSizeTest := make([][]string, 25) // More than 2 batches worth
	for i := range 25 {
		batchSizeTest[i] = []string{
			"Test Merchant", "TM", "logo.png", "test@merchant.com", "08123456789",
			"ID", "INDIVIDUAL", "PT", "John Doe", "08987654321", "john@test.com",
			"Jl Test 123", "12345", "1234567890", "BCA",
		}
	}

	request := &merchantModel.BulkCreateSubMerchantRequest{
		MerchantId:         "parent-merchant-123",
		KYCType:            constant.MerchantKYCTypeKYC,
		FileName:           "large-batch.csv",
		SubmerchantDetails: batchSizeTest,
	}

	t.Run("SUCCESS: Batch processing with multiple batches", func(t *testing.T) {
		// Initialize mocks
		repo := mockRepo.NewIMerchantRepository(t)
		logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
		userRepo := mockRepo.NewIUserRepository(t)
		jwtSvc := mockJWT.NewIJwt(t)
		encryptMock := mockEncrypt.NewICrypto(t)
		rabbitMqExt := mockRabbitMQ.NewRabbitMQExt(t)
		redisExt := mockRedis.NewIRedisExt(t)
		validator := validatorExt.New()

		// Parent merchant validation
		repo.On("FindMerchantByID", mock.Anything, "parent-merchant-123").Return(&merchantModel.Merchant{
			UUID: "parent-merchant-123",
			Name: "Parent Merchant",
		}, nil)

		// Mock RabbitMQ publish for multiple batches (25 items with batch size 10 = 3 batches)
		rabbitMqExt.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil).Times(3)

		// Create service instance
		s := New(
			repo, logger, userRepo, jwtSvc, rabbitMqExt, encryptMock,
			WithRedisClient(redisExt),
			WithValidator(validator),
		)

		// Execute the function
		result, err := s.BulkCreateSubMerchant(ctx, request)

		// Assertions
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.TotalFailed) // No validation failures expected
		assert.NotEmpty(t, result.ID)          // Session ID should be generated
		assert.Empty(t, result.Results)        // No direct results since processing is asynchronous via RabbitMQ

		// Assert mock expectations
		repo.AssertExpectations(t)
		rabbitMqExt.AssertExpectations(t)
	})
}

func TestProcessBulkCreateSubMerchant(t *testing.T) {
	ctxValue := context.WithValue(t.Context(), mySqlExt.CtxSqlTx, &sqlx.Tx{})

	vaultTransit := mockVault.NewIVaultTransit(t)
	vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Return(&vault.EncryptResponse{}, nil)

	tests := []struct {
		name          string
		request       *merchantModel.ProcessBulkCreateSubMerchantRequest
		setupMocks    func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, beneficiaryAccountSvc *mockService.IBeneficiaryAccountService, userSvc *mockService.IUserService, bankAccountRepo *mockRepo.IBankAccountRepository, redisClientMock redismock.ClientMock)
		expectedError bool
	}{
		{
			name: "SUCCESS: Process single submerchant",
			request: &merchantModel.ProcessBulkCreateSubMerchantRequest{
				ID:         "test-session-id",
				MerchantId: "parent-merchant-123",
				KYCType:    constant.MerchantKYCTypeKYC,
				Batch:      0,
				FileName:   "test-file.csv",
				SubmerchantDetails: []merchantModel.BulkSubMerchantDetailRequest{
					{
						Row: 0,
						Detail: merchantModel.MerchantRequest{
							Name:              "Test Merchant",
							ShortName:         "TM",
							Logo:              "logo.png",
							MerchantEmail:     "test@merchant.com",
							MerchantPhone:     "08123456789",
							BusinessCountry:   "ID",
							BusinessType:      "INDIVIDUAL",
							BusinessStructure: "PT",
							PICName:           "John Doe",
							PICPhone:          "08987654321",
							PICEmail:          "john@test.com",
							Address:           "Jl Test 123",
							PostCode:          "12345",
							BankAccount: &merchantModel.MerchantBankAccountRequest{
								AccountNumber: "1234567890",
								ChannelCode:   "BCA",
							},
							KYCStatus:      constant.KYCStatusApproved,
							PICInvitation:  true,
							MerchantStatus: constant.MerchantStatusActive,
							ParentID:       "parent-merchant-123",
							RequesterID:    "parent-merchant-123",
						},
					},
				},
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, beneficiaryAccountSvc *mockService.IBeneficiaryAccountService, userSvc *mockService.IUserService, bankAccountRepo *mockRepo.IBankAccountRepository, redisClientMock redismock.ClientMock) {
				// Mock CreateSubMerchant success
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{
					UUID: "requester-merchant-id",
					Name: "Requester Merchant",
				}, nil)

				userSvc.On("FindUserByEmail", mock.Anything, "john@test.com").Return(nil, nil)

				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Return(&beneficiaryAccountModel.Account{
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryAccountName: "Test Account",
					BeneficiaryBankCode:    "014",
					BeneficiaryBankName:    "BCA",
				}, nil)

				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)

				accountService.On("CreateMerchantAccount", mock.Anything, mock.Anything, constant.UserTypeSubMerchant).Return(nil)

				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("test-secret")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("encrypted-key", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)

				bankAccountRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				userSvc.On("CreateMerchantUser", mock.Anything, mock.Anything).Return(nil, nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
				repo.On("RollbackTransaction", mock.Anything).Return(nil).Maybe()

				redisClientMock.ExpectExists(constant.MerchantReservedShortNameKey).SetVal(1)
				redisClientMock.ExpectHGet(constant.MerchantReservedShortNameKey, "TM").SetErr(redisExt.ErrNil)

				// Mock Redis ZAdd for storing results - one for result data, one for filename
				redisClientMock.MatchExpectationsInOrder(false)
				redisClientMock.ExpectZAdd(mock.Anything).SetVal(1)
				redisClientMock.ExpectZAdd(mock.Anything).SetVal(1)
				redisClientMock.ExpectExpire(mock.Anything, constant.MerchantBulkCreateSubMerchantSessionIDCacheTTL).SetVal(true)
			},
			expectedError: false,
		},
		{
			name: "ERROR: CreateSubMerchant failure",
			request: &merchantModel.ProcessBulkCreateSubMerchantRequest{
				ID:         "test-session-id",
				MerchantId: "parent-merchant-123",
				KYCType:    constant.MerchantKYCTypeKYC,
				Batch:      0,
				FileName:   "test-file.csv",
				SubmerchantDetails: []merchantModel.BulkSubMerchantDetailRequest{
					{
						Row: 0,
						Detail: merchantModel.MerchantRequest{
							Name:              "Test Merchant",
							ShortName:         "TM",
							Logo:              "logo.png",
							MerchantEmail:     "test@merchant.com",
							MerchantPhone:     "08123456789",
							BusinessCountry:   "ID",
							BusinessType:      "INDIVIDUAL",
							BusinessStructure: "PT",
							PICName:           "John Doe",
							PICPhone:          "08987654321",
							PICEmail:          "john@test.com",
							Address:           "Jl Test 123",
							PostCode:          "12345",
							BankAccount: &merchantModel.MerchantBankAccountRequest{
								AccountNumber: "1234567890",
								ChannelCode:   "BCA",
							},
							KYCStatus:      constant.KYCStatusApproved,
							PICInvitation:  false,
							MerchantStatus: constant.MerchantStatusActive,
							ParentID:       "parent-merchant-123",
							RequesterID:    "parent-merchant-123",
						},
					},
				},
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, beneficiaryAccountSvc *mockService.IBeneficiaryAccountService, userSvc *mockService.IUserService, bankAccountRepo *mockRepo.IBankAccountRepository, redisClientMock redismock.ClientMock) {
				// Mock CreateSubMerchant failure
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, errors.New("merchant not found"))

				// Mock Redis ZAdd for storing error results
				redisClientMock.MatchExpectationsInOrder(false)
				redisClientMock.ExpectZAdd(mock.Anything).SetVal(1)
				redisClientMock.ExpectZAdd(mock.Anything).SetVal(1)
				redisClientMock.ExpectExpire(mock.Anything, constant.MerchantBulkCreateSubMerchantSessionIDCacheTTL).SetVal(true)
			},
			expectedError: false, // Function doesn't return error, just logs and stores results
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			repo := mockRepo.NewIMerchantRepository(t)
			accountService := mockService.NewIAccountService(t)
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			encryptMock := mockEncrypt.NewICrypto(t)
			beneficiaryAccountSvc := mockService.NewIBeneficiaryAccountService(t)
			userSvc := mockService.NewIUserService(t)
			bankAccountRepo := mockRepo.NewIBankAccountRepository(t)
			userRepo := mockRepo.NewIUserRepository(t)
			jwtSvc := mockJWT.NewIJwt(t)
			rabbitMqExt := mockRabbitMQ.NewRabbitMQExt(t)
			validator := validatorExt.New()
			redisClient, redisClientMock := redismock.NewClientMock()

			// Setup mocks
			tt.setupMocks(repo, accountService, encryptMock, beneficiaryAccountSvc, userSvc, bankAccountRepo, redisClientMock)

			// Create service instance
			s := New(
				repo, logger, userRepo, jwtSvc, rabbitMqExt, encryptMock,
				WithRedisClient(redisExt.WrapRedisClient(redisClient, nil)),
				WithValidator(validator),
				WithAccountService(accountService),
				WithBeneficiaryAccountService(beneficiaryAccountSvc),
				WithBankAccountRepository(bankAccountRepo),
				WithUserService(userSvc),
				WithServiceConfig(&config.Config{}),
				WithVaultTransit(vaultTransit),
			)

			// Execute the function
			err := s.ProcessBulkCreateSubMerchant(t.Context(), tt.request)

			// Assertions
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			// Assert mock expectations
			repo.AssertExpectations(t)
			accountService.AssertExpectations(t)
			encryptMock.AssertExpectations(t)
			beneficiaryAccountSvc.AssertExpectations(t)
			userSvc.AssertExpectations(t)
			bankAccountRepo.AssertExpectations(t)
		})
	}
}

func TestGetBulkCreateSubMerchantSummary(t *testing.T) {
	ctx := context.Background()
	sesionId := "test-session-id"
	merchantId := "merchant-123"
	redisKey := fmt.Sprintf(constant.MerchantBulkCreateSubMerchantSessionIDCacheKey, merchantId, sesionId)

	tests := []struct {
		name           string
		request        *merchantModel.GetBulkCreateSubMerchantSummaryRequest
		setupMocks     func(redisMock redismock.ClientMock)
		expectedError  bool
		expectedResult *merchantModel.BulkCreateSubMerchantResponse
	}{
		{
			name: "SUCCESS: Get summary with results",
			request: &merchantModel.GetBulkCreateSubMerchantSummaryRequest{
				ID:         "test-session-id",
				MerchantId: "merchant-123",
			},
			setupMocks: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZRange(redisKey, 0, 1).SetVal([]string{"test-file.csv"})

				// Mock results retrieval
				resultData1, _ := json.Marshal(merchantModel.BulkCreateSubMerchantDetailResponse{
					Row:          0,
					MerchantID:   "merchant-id-1",
					MerchantName: "Merchant 1",
				})
				resultData2, _ := json.Marshal(merchantModel.BulkCreateSubMerchantDetailResponse{
					Row:   1,
					Error: "validation error",
				})
				redisMock.ExpectZRange(redisKey, 1, -1).SetVal([]string{string(resultData1), string(resultData2)})
			},
			expectedError: false,
			expectedResult: &merchantModel.BulkCreateSubMerchantResponse{
				ID:           "test-session-id",
				FileName:     "test-file.csv",
				TotalSuccess: 1,
				TotalFailed:  1,
				Results: []merchantModel.BulkCreateSubMerchantDetailResponse{
					{
						Row:          0,
						MerchantID:   "merchant-id-1",
						MerchantName: "Merchant 1",
					},
					{
						Row:   1,
						Error: "validation error",
					},
				},
			},
		},
		{
			name: "ERROR: Redis error when retrieving file name",
			request: &merchantModel.GetBulkCreateSubMerchantSummaryRequest{
				ID:         "test-session-id",
				MerchantId: "merchant-123",
			},
			setupMocks: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZRange(redisKey, 0, 1).SetErr(errors.New("redis connection error"))

			},
			expectedError: true,
		},
		{
			name: "ERROR: Get results from redis",
			request: &merchantModel.GetBulkCreateSubMerchantSummaryRequest{
				ID:         "test-session-id",
				MerchantId: "merchant-123",
			},
			setupMocks: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZRange(redisKey, 0, 1).SetVal([]string{"test-file.csv"})
				redisMock.ExpectZRange(redisKey, 1, -1).SetErr(errors.New("redis connection error"))
			},
			expectedError: true,
		},
		{
			name: "ERROR: Unmarshal result from redis",
			request: &merchantModel.GetBulkCreateSubMerchantSummaryRequest{
				ID:         "test-session-id",
				MerchantId: "merchant-123",
			},
			setupMocks: func(redisMock redismock.ClientMock) {
				redisMock.ExpectZRange(redisKey, 0, 1).SetVal([]string{"test-file.csv"})

				// Mock results retrieval
				resultData1, _ := json.Marshal(merchantModel.BulkCreateSubMerchantDetailResponse{
					Row:          0,
					MerchantID:   "merchant-id-1",
					MerchantName: "Merchant 1",
				})
				redisMock.ExpectZRange(redisKey, 1, -1).SetVal([]string{string(resultData1), "{}/{}"})
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			repo := mockRepo.NewIMerchantRepository(t)
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			encryptMock := mockEncrypt.NewICrypto(t)
			userRepo := mockRepo.NewIUserRepository(t)
			jwtSvc := mockJWT.NewIJwt(t)
			rabbitMqExt := mockRabbitMQ.NewRabbitMQExt(t)
			validator := validatorExt.New()

			redisClient, redisClientMock := redismock.NewClientMock()

			// Setup mocks
			tt.setupMocks(redisClientMock)

			// Create service instance
			s := New(
				repo, logger, userRepo, jwtSvc, rabbitMqExt, encryptMock,
				WithRedisClient(redisExt.WrapRedisClient(redisClient, nil)),
				WithValidator(validator),
			)

			// Execute the function
			result, err := s.GetBulkCreateSubMerchantSummary(ctx, tt.request)

			// Assertions
			if tt.expectedError {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)

				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult.ID, result.ID)
					assert.Equal(t, tt.expectedResult.FileName, result.FileName)
					assert.Equal(t, tt.expectedResult.TotalSuccess, result.TotalSuccess)
					assert.Equal(t, tt.expectedResult.TotalFailed, result.TotalFailed)
					assert.Equal(t, len(tt.expectedResult.Results), len(result.Results))
				}
			}

		})
	}
}

// Note: storeBulkCreateSubMerchantResult is a private method and is tested indirectly
// through ProcessBulkCreateSubMerchant function. In a real test environment, you would
// verify Redis interactions through integration tests or by testing the public methods
// that call this private function.
