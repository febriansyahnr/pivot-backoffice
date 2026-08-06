package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/location"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	mockRabbitMQ "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/jmoiron/sqlx"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSubMerchant(t *testing.T) {

	request := &merchantModel.MerchantRequest{
		Name:        "test",
		ParentID:    "test",
		ShortName:   "sample",
		DistrictId:  1,
		BankAccount: &merchantModel.MerchantBankAccountRequest{},
		PICEmail:    "john.doe@email.com",
	}
	bankAccountRepo := mockRepo.NewIBankAccountRepository(t)
	beneficiaryAccountSvc := mockService.NewIBeneficiaryAccountService(t)
	userSvc := mockService.NewIUserService(t)

	ctxValue := context.WithValue(t.Context(), mySqlExt.CtxSqlTx, &sqlx.Tx{})

	vaultTransit := vaultMock.NewIVaultTransit(t)

	tests := []struct {
		name          string
		setupMocks    func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, redis *redisExtMock.IRedisExt)
		expectedError bool
		request       *merchantModel.MerchantRequest
	}{
		{
			name: "SUCCESS: Creation",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Times(2).Return(&vault.EncryptResponse{}, nil)

				// Create merchant account
				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("nil", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)

				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&beneficiaryAccountModel.Account{}, nil)
				bankAccountRepo.On("Create", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil)
			},
			expectedError: false,
			request:       request,
		},
		{
			name: "SUCCESS: Create without District Id",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Times(2).Return(&vault.EncryptResponse{}, nil)

				// Create merchant account
				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("nil", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)

				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&beneficiaryAccountModel.Account{}, nil)
				bankAccountRepo.On("Create", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil)
			},
			expectedError: false,
			request: &merchantModel.MerchantRequest{
				Name:        "test",
				ParentID:    "test",
				BankAccount: &merchantModel.MerchantBankAccountRequest{},
				ShortName:   "sample",
			},
		},
		{
			name: "error get merchant by id",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error get user by email address",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				request.PICInvitation = true

				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil)
				userSvc.On("FindUserByEmail", mock.Anything, request.PICEmail).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error pic email already registered",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil)
				userSvc.On("FindUserByEmail", mock.Anything, request.PICEmail).Once().Return(&user.User{}, nil)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error get district by id",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				request.PICInvitation = false

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error district not found",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(nil, nil)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error bank channel code not found",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
			},
			expectedError: true,
			request: &merchantModel.MerchantRequest{
				BankAccount: &merchantModel.MerchantBankAccountRequest{
					ChannelCode: "XXXX",
				},
			},
		},
		{
			name: "error get country",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, mock.Anything).Return(&location.District{}, nil)
				countrySvc.On("FindByCode", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
			expectedError: true,
			request: &merchantModel.MerchantRequest{
				Name:            "test",
				ParentID:        "test",
				DistrictId:      1,
				BankAccount:     &merchantModel.MerchantBankAccountRequest{},
				PICEmail:        "john.doe@email.com",
				CountryOfEntity: "ID",
			},
		},
		{
			name: "country not found",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, mock.Anything).Return(&location.District{}, nil)
				countrySvc.On("FindByCode", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedError: true,
			request: &merchantModel.MerchantRequest{
				Name:            "test",
				ParentID:        "test",
				DistrictId:      1,
				BankAccount:     &merchantModel.MerchantBankAccountRequest{},
				PICEmail:        "john.doe@email.com",
				CountryOfEntity: "ID",
			},
		},
		{
			name: "error bank account not found",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, mock.Anything).Return(&location.District{}, nil)
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil, constant.ErrInvalidAccount)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error find bank account",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, mock.Anything).Return(&location.District{}, nil)
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error encrypting callback api key",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, mock.Anything).Return(&location.District{}, nil)
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", mock.Anything, mock.Anything).Once().Return(&beneficiaryAccountModel.Account{}, nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error begin transaction",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Once().Return(&beneficiaryAccountModel.Account{}, nil)
				repo.On("BeginTransaction", mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Return(&vault.EncryptResponse{}, nil)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error reserved shortname",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Run(func(args mock.Arguments) {
					val := args.Get(3).(*string)
					*val = ""
				}).Return(nil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Once().Return(&beneficiaryAccountModel.Account{}, nil)
			},
			expectedError: true,
			request: &merchantModel.MerchantRequest{
				Name:        "test",
				ParentID:    "test",
				ShortName:   "sample",
				DistrictId:  1,
				BankAccount: &merchantModel.MerchantBankAccountRequest{},
				PICEmail:    "john.doe@email.com",
			},
		},
		{
			name: "error creating new submerchant",
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockService.IAccountService, _ *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Once().Return(&beneficiaryAccountModel.Account{}, nil)
				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("RollbackTransaction", mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(errors.New("create error"))

			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error creating new account",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Once().Return(&beneficiaryAccountModel.Account{}, nil)
				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("RollbackTransaction", mock.Anything).Return(nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, mock.Anything).Return(errors.New("create error"))
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error creating new merchant auth",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Once().Return(&beneficiaryAccountModel.Account{}, nil)
				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)
				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("nil", errors.New("errors"))
				repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error create bank account",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&beneficiaryAccountModel.Account{}, nil)

				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)

				// Create merchant account
				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("nil", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)
				bankAccountRepo.On("Create", constant.ValueCtxMockType(), mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)

				repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error create sub-merchant user",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				request.PICInvitation = true

				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				userSvc.On("FindUserByEmail", mock.Anything, request.PICEmail).Once().Return(nil, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&beneficiaryAccountModel.Account{}, nil)

				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)

				// Create merchant account
				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("nil", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)
				bankAccountRepo.On("Create", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil)
				userSvc.On("CreateMerchantUser", mock.Anything, mock.Anything).Once().Return(nil, constant.ErrSomeErrorForUnitTest)

				repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error commit transaction",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				request.PICInvitation = false

				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&beneficiaryAccountModel.Account{}, nil)

				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)

				// Create merchant account
				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("nil", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)
				bankAccountRepo.On("Create", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil)

				repo.On("CommitTransaction", mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
				repo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error: when sub-merchant was NON_KYC and create sub-merchant, then should return error",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{
					ParentID: sql.NullString{
						String: "parent-id",
						Valid:  true,
					},
					KYCStatus: sql.NullString{
						String: constant.KYCStatusNotRequired,
						Valid:  true,
					},
				}, nil)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "error: when sub-merchant was IN_REVIEW and create sub-merchant, then should return error",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{
					ParentID: sql.NullString{
						String: "parent-id",
						Valid:  true,
					},
					KYCStatus: sql.NullString{
						String: constant.KYCStatusInReview,
						Valid:  true,
					},
				}, nil)
			},
			expectedError: true,
			request:       request,
		},
		{
			name: "success: when sub-merchant was KYC, already approved and create sub-merchant, then should not return error",
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, countrySvc *mockService.ICountryService, mockRedis *redisExtMock.IRedisExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "SAMPLE", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{
					ParentID: sql.NullString{
						String: "parent-id",
						Valid:  true,
					},
					KYCStatus: sql.NullString{
						String: constant.KYCStatusApproved,
						Valid:  true,
					},
				}, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)

				// Create merchant account
				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("nil", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)

				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("CommitTransaction", mock.Anything).Return(nil)
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&beneficiaryAccountModel.Account{}, nil)
				bankAccountRepo.On("Create", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil)
			},
			expectedError: false,
			request:       request,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockRepo.NewIMerchantRepository(t)
			accountService := mockService.NewIAccountService(t)
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			encryptMock := mockEncrypt.NewICrypto(t)
			locRepo := mockRepo.NewIAddrLocationRepository(t)
			countrySvc := mockService.NewICountryService(t)
			redisClient := redisExtMock.NewIRedisExt(t)
			gcs := gcsMock.NewGCSService(t)
			excel := xlsxMock.NewExceler(t)
			cfg := &config.Config{
				GCSConfig: config.GCSConfig{
					MerchantReservedSortName: "reserved-names",
					ServiceBucketName:        "test-bucket",
				},
			}

			tt.setupMocks(repo, accountService, encryptMock, locRepo, countrySvc, redisClient)

			// Service
			s := New(
				repo, logger, nil, nil, nil, encryptMock,
				WithAccountService(accountService),
				WithLocationRepository(locRepo),
				WithBeneficiaryAccountService(beneficiaryAccountSvc),
				WithBankAccountRepository(bankAccountRepo),
				WithUserService(userSvc),
				WithCountryService(countrySvc),
				WithServiceConfig(cfg),
				WithGCSService(gcs),
				WithExcelLibrary(excel),
				WithRedisClient(redisClient),
				WithVaultTransit(vaultTransit),
			)

			// Execute
			_, err := s.CreateSubMerchant(t.Context(), tt.request)

			if tt.expectedError {
				assert.NotNil(t, err)

			} else {
				assert.Nil(t, err)
			}
			_ = userSvc.AssertExpectations(t)
			_ = repo.AssertExpectations(t)
			_ = accountService.AssertExpectations(t)
			_ = encryptMock.AssertExpectations(t)
			_ = locRepo.AssertExpectations(t)
			_ = beneficiaryAccountSvc.AssertExpectations(t)
			_ = bankAccountRepo.AssertExpectations(t)
		})
	}
}

func TestCreateSubMerchantAutoApproveKYCInStaging(t *testing.T) {
	ctxValue := context.WithValue(t.Context(), mySqlExt.CtxSqlTx, &sqlx.Tx{})

	vaultTransit := vaultMock.NewIVaultTransit(t)
	vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Return(&vault.EncryptResponse{}, nil)

	// Initialize shared mocks
	bankAccountRepo := mockRepo.NewIBankAccountRepository(t)
	beneficiaryAccountSvc := mockService.NewIBeneficiaryAccountService(t)
	userSvc := mockService.NewIUserService(t)
	userRepo := mockRepo.NewIUserRepository(t)

	tests := []struct {
		name          string
		request       *merchantModel.MerchantRequest
		config        *config.Config
		setupMocks    func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, paymentMethodRepo *mockRepo.IPaymentMethodRepository, qrisSvc *mockService.IQrisService, redis *redisExtMock.IRedisExt, rabbitMq *mockRabbitMQ.RabbitMQExt)
		expectedError bool
		checkResult   func(t *testing.T, merchant *merchantModel.Merchant)
	}{
		{
			name: "SUCCESS: Auto-approve KYC in staging environment",
			request: &merchantModel.MerchantRequest{
				Name:           "Test Merchant",
				ParentID:       "parent-123",
				ShortName:      "TEST",
				KYCStatus:      constant.KYCStatusWaitingForDocument,
				MerchantStatus: constant.MerchantStatusCreated,
				BankAccount:    &merchantModel.MerchantBankAccountRequest{AccountNumber: "123", ChannelCode: "BCA"},
				PICEmail:       "test@test.com",
			},
			config: &config.Config{
				Environment: constant.EnvironmentStaging,
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, paymentMethodRepo *mockRepo.IPaymentMethodRepository, qrisSvc *mockService.IQrisService, mockRedis *redisExtMock.IRedisExt, rabbitMq *mockRabbitMQ.RabbitMQExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "TEST", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("GenerateNewMID", constant.ValueCtxMockType()).Return("valid-mid", nil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)

				// Verify that Create is called with ACTIVE status and APPROVED KYC (auto-approved in staging)
				repo.On("Create", constant.ValueCtxMockType(), mock.MatchedBy(func(m *merchantModel.Merchant) bool {
					// Status should be auto-approved to ACTIVE before Create is called
					return m.Status == constant.MerchantStatusActive &&
						m.KYCStatus.Valid &&
						m.KYCStatus.String == constant.KYCStatusApproved
				})).Return(nil)

				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)

				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("test-secret")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("encrypted", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)

				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Return(&beneficiaryAccountModel.Account{
					BeneficiaryAccountNo:   "123",
					BeneficiaryAccountName: "Test",
					BeneficiaryBankCode:    "014",
					BeneficiaryBankName:    "BCA",
				}, nil)
				bankAccountRepo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)

				repo.On("CommitTransaction", mock.Anything).Return(nil)

				// Mock EnableAllPaymentMethod - it will be called after commit
				paymentMethodRepo.On("GetAllPaymentMethodByCategory", mock.Anything, constant.TypePayment).Return([]*paymentModel.PaymentMethod{}, nil).Maybe()

				// Mock callbacks - should send 2 callbacks in staging (PENDING and APPROVED)
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Times(2)
			},
			expectedError: false,
			checkResult: func(t *testing.T, merchant *merchantModel.Merchant) {
				assert.Equal(t, constant.MerchantStatusActive, merchant.Status)
				assert.Equal(t, constant.KYCStatusApproved, merchant.KYCStatus.String)
				assert.True(t, merchant.KYCStatus.Valid)
			},
		},
		{
			name: "SUCCESS: No auto-approve in production environment",
			request: &merchantModel.MerchantRequest{
				Name:           "Test Merchant",
				ParentID:       "parent-123",
				ShortName:      "TEST",
				KYCStatus:      constant.KYCStatusWaitingForDocument,
				MerchantStatus: constant.MerchantStatusCreated,
				BankAccount:    &merchantModel.MerchantBankAccountRequest{AccountNumber: "123", ChannelCode: "BCA"},
				PICEmail:       "test@test.com",
			},
			config: &config.Config{
				Environment: constant.EnvironmentProduction,
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, paymentMethodRepo *mockRepo.IPaymentMethodRepository, qrisSvc *mockService.IQrisService, mockRedis *redisExtMock.IRedisExt, rabbitMq *mockRabbitMQ.RabbitMQExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "TEST", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)

				// Verify that Create is called with CREATED status (NOT auto-approved in production)
				repo.On("Create", constant.ValueCtxMockType(), mock.MatchedBy(func(m *merchantModel.Merchant) bool {
					// Status should remain CREATED in production
					return m.Status == constant.MerchantStatusCreated &&
						m.KYCStatus.Valid &&
						m.KYCStatus.String == constant.KYCStatusWaitingForDocument
				})).Return(nil)

				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)

				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("test-secret")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("encrypted", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)

				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Return(&beneficiaryAccountModel.Account{
					BeneficiaryAccountNo:   "123",
					BeneficiaryAccountName: "Test",
					BeneficiaryBankCode:    "014",
					BeneficiaryBankName:    "BCA",
				}, nil)
				bankAccountRepo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)

				// Should NOT call UpdateKYC in production
				// Should NOT call EnableAllPaymentMethod in production
				// Should NOT send callbacks in production

				repo.On("CommitTransaction", mock.Anything).Return(nil)
			},
			expectedError: false,
			checkResult: func(t *testing.T, merchant *merchantModel.Merchant) {
				// Status should remain as Created (not auto-approved)
				assert.Equal(t, constant.MerchantStatusCreated, merchant.Status)
			},
		},
		{
			name: "SUCCESS: No auto-approve when KYC status is already approved",
			request: &merchantModel.MerchantRequest{
				Name:           "Test Merchant",
				ParentID:       "parent-123",
				ShortName:      "TEST",
				KYCStatus:      constant.KYCStatusApproved,
				MerchantStatus: constant.MerchantStatusActive,
				BankAccount:    &merchantModel.MerchantBankAccountRequest{AccountNumber: "123", ChannelCode: "BCA"},
				PICEmail:       "test@test.com",
			},
			config: &config.Config{
				Environment: constant.EnvironmentStaging,
			},
			setupMocks: func(repo *mockRepo.IMerchantRepository, accountService *mockService.IAccountService, crypto *mockEncrypt.ICrypto, locRepo *mockRepo.IAddrLocationRepository, paymentMethodRepo *mockRepo.IPaymentMethodRepository, qrisSvc *mockService.IQrisService, mockRedis *redisExtMock.IRedisExt, rabbitMq *mockRabbitMQ.RabbitMQExt) {
				mockRedis.On("Exists", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				mockRedis.On("HGetScan", constant.ValueCtxMockType(), constant.MerchantReservedShortNameKey, "TEST", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()

				repo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				repo.On("BeginTransaction", mock.Anything).Return(ctxValue, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)

				accountService.On("CreateMerchantAccount", constant.ValueCtxMockType(), mock.Anything, constant.UserTypeSubMerchant).Return(nil)

				crypto.On("GenerateRandomPKCS8Key").Return(nil, nil)
				crypto.On("SecretKeyFromUUID", mock.Anything).Return("test-secret")
				crypto.On("Encrypt", mock.Anything, mock.Anything).Return("encrypted", nil)
				repo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)

				beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", constant.ValueCtxMockType(), mock.Anything).Return(&beneficiaryAccountModel.Account{
					BeneficiaryAccountNo:   "123",
					BeneficiaryAccountName: "Test",
					BeneficiaryBankCode:    "014",
					BeneficiaryBankName:    "BCA",
				}, nil)
				bankAccountRepo.On("Create", constant.ValueCtxMockType(), mock.Anything).Return(nil)

				// Should NOT call UpdateKYC because status is already Approved

				repo.On("CommitTransaction", mock.Anything).Return(nil)

				// Should call EnableAllPaymentMethod because merchant is Active
				paymentMethodRepo.On("GetAllPaymentMethodByCategory", mock.Anything, constant.TypePayment).Return([]*paymentModel.PaymentMethod{}, nil).Maybe()

				// Should send callbacks because merchant is ACTIVE in staging
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil).Times(2)
			},
			expectedError: false,
			checkResult: func(t *testing.T, merchant *merchantModel.Merchant) {
				assert.Equal(t, constant.MerchantStatusActive, merchant.Status)
				assert.Equal(t, constant.KYCStatusApproved, merchant.KYCStatus.String)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			repo := mockRepo.NewIMerchantRepository(t)
			accountService := mockService.NewIAccountService(t)
			encryptMock := mockEncrypt.NewICrypto(t)
			locRepo := mockRepo.NewIAddrLocationRepository(t)
			paymentMethodRepo := mockRepo.NewIPaymentMethodRepository(t)
			qrisSvc := mockService.NewIQrisService(t)
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			redisClient := redisExtMock.NewIRedisExt(t)
			gcs := gcsMock.NewGCSService(t)
			excel := xlsxMock.NewExceler(t)
			rabbitMq := mockRabbitMQ.NewRabbitMQExt(t)

			// Setup mocks
			tt.setupMocks(repo, accountService, encryptMock, locRepo, paymentMethodRepo, qrisSvc, redisClient, rabbitMq)

			// Create service with config
			s := New(
				repo,
				logger,
				userRepo,
				nil,
				rabbitMq,
				encryptMock,
				WithAccountService(accountService),
				WithLocationRepository(locRepo),
				WithBeneficiaryAccountService(beneficiaryAccountSvc),
				WithBankAccountRepository(bankAccountRepo),
				WithUserService(userSvc),
				WithServiceConfig(tt.config),
				WithPaymentMethodRepository(paymentMethodRepo),
				WithQrisService(qrisSvc),
				WithGCSService(gcs),
				WithExcelLibrary(excel),
				WithRedisClient(redisClient),
				WithVaultTransit(vaultTransit),
			)

			// Execute
			result, err := s.CreateSubMerchant(t.Context(), tt.request)

			// Assert
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			// Assert expectations
			repo.AssertExpectations(t)
			accountService.AssertExpectations(t)
			encryptMock.AssertExpectations(t)
			beneficiaryAccountSvc.AssertExpectations(t)
			bankAccountRepo.AssertExpectations(t)
			paymentMethodRepo.AssertExpectations(t)
		})
	}
}
