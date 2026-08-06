package merchant

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/location"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	vaultTransit := vaultMock.NewIVaultTransit(t)
	locRepo := repositoryMocks.NewIAddrLocationRepository(t)

	encryptResults := make([]vault.EncryptResponse, 3)
	expectedUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "pass",
		Password:   "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1",
		MerchantId: "",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		CreatedAt:  time.Now(),
	}

	createdMerchant := &merchantModel.Merchant{
		UUID:       "merchant-id",
		Status:     constant.MerchantStatusActive,
		Name:       "test",
		Logo:       "https://paper.id/test.jpg",
		MID:        sql.NullString{String: "0000", Valid: true},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IndustryID: uuid.NewString(),
		DistrictId: 123,
	}

	testCases := []struct {
		name          string
		input         *merchantModel.Merchant
		expectedUser  *userModel.User
		expectedError string
		mocksSetup    func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService)
		wantErr       bool
	}{
		{
			name:  "ERROR: err generate MID",
			input: createdMerchant,
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("", errors.New("some-error"))

			},
			wantErr: true,
		},
		{
			name:  "SUCCESS: successfully create merchant",
			input: createdMerchant,
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*merchant.Merchant"),
				).Return(nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)

				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("0000", nil)
				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(nil)
				trxRepo.On(
					"CreateMerchantAuth",
					mock.Anything,
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(nil)

				mockString := "mockString"
				enc.On("GenerateRandomPKCS8Key").Return([]byte{}, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return(mockString)
				enc.On("Encrypt", mock.Anything, mock.Anything).Return(mockString, nil)

				rmq.On(
					"Publish",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				accountRepo.On(
					"Create",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)
				vaultTransit.On("BatchEncrypt", mock.Anything, mock.Anything).Once().Return(encryptResults, nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: error find user by id",
			input: createdMerchant,
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("error when find user by id"))
			},
			wantErr: true,
		},
		{
			name:          "ERROR: user not found",
			input:         createdMerchant,
			expectedUser:  nil,
			expectedError: "ERROR_NOT_FOUND",
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: user already have merchant",
			input: createdMerchant,
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = "merchant-id"
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Get district by id",
			input: createdMerchant,
			mocksSetup: func(_ *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, _ *mockRabbitMq.RabbitMQExt, _ *mockEncrypt.ICrypto, _ *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = ""
				userRepo.On("FindUserByID", mock.Anything, constant.StringMockType()).Return(expectedUser, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: District id not found",
			input: createdMerchant,
			mocksSetup: func(_ *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, _ *mockRabbitMq.RabbitMQExt, _ *mockEncrypt.ICrypto, _ *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				userRepo.On("FindUserByID", mock.Anything, constant.StringMockType()).Return(expectedUser, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Encrypt merchant secrets",
			input: createdMerchant,
			mocksSetup: func(merchantRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, _ *mockRabbitMq.RabbitMQExt, _ *mockEncrypt.ICrypto, _ *repositoryMocks.IAccountRepository, _ *mocks.IIndustryService, _ *mocks.ICountryService) {
				expectedUser.MerchantId = ""
				userRepo.On("FindUserByID", mock.Anything, mock.Anything).Return(expectedUser, nil)
				locRepo.On("GetDistrictById", mock.Anything, mock.Anything).Return(&location.District{}, nil)
				merchantRepo.On("GenerateNewMID", mock.Anything).Return("0000", nil)
				vaultTransit.On("BatchEncrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Invalid encrypt results",
			input: createdMerchant,
			mocksSetup: func(merchantRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, _ *mockRabbitMq.RabbitMQExt, _ *mockEncrypt.ICrypto, _ *repositoryMocks.IAccountRepository, _ *mocks.IIndustryService, _ *mocks.ICountryService) {
				expectedUser.MerchantId = ""
				userRepo.On("FindUserByID", mock.Anything, mock.Anything).Return(expectedUser, nil)
				locRepo.On("GetDistrictById", mock.Anything, mock.Anything).Return(&location.District{}, nil)
				merchantRepo.On("GenerateNewMID", mock.Anything).Return("0000", nil)
				vaultTransit.On("BatchEncrypt", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: error create merchant",
			input: createdMerchant,
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = ""
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)

				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("0000", nil)
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*merchant.Merchant"),
				).Return(errors.New("error when create merchant"))
				vaultTransit.On("BatchEncrypt", mock.Anything, mock.Anything).Return(encryptResults, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: error update user",
			input: createdMerchant,
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)

				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("0000", nil)
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*merchant.Merchant"),
				).Return(nil)
				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(errors.New("error when update user"))
			},
			wantErr: true,
		},
		{
			name:  "ERROR: error GenerateRandomPKCS8Key",
			input: createdMerchant,
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = ""
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)

				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*merchant.Merchant"),
				).Return(nil)
				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("0000", nil)
				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(nil)

				enc.On("GenerateRandomPKCS8Key").Return(nil, errors.New("error when GenerateRandomPKCS8Key"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: encrypt",

			input: createdMerchant,
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = ""

				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*merchant.Merchant"),
				).Return(nil)
				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("0000", nil)
				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(nil)

				mockString := "mockString"
				enc.On("GenerateRandomPKCS8Key").Return([]byte{}, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return(mockString)
				enc.On("Encrypt", mock.Anything, mock.Anything).Return("", errors.New("error when encrypt"))

			},
			wantErr: true,
		},
		{
			name:  "ERROR: create account failed",
			input: createdMerchant,
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = ""

				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*merchant.Merchant"),
				).Return(nil)
				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("0000", nil)
				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(nil)
				trxRepo.On(
					"CreateMerchantAuth",
					mock.Anything,
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(nil)

				mockString := "mockString"
				enc.On("GenerateRandomPKCS8Key").Return([]byte{}, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return(mockString)
				enc.On("Encrypt", mock.Anything, mock.Anything).Return(mockString, nil)

				accountRepo.On(
					"Create",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(errors.New("error when create account for merchant"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: invalid risk level",
			input: &merchantModel.Merchant{
				UUID:       "merchant-id",
				Status:     constant.MerchantStatusActive,
				Name:       "test",
				Logo:       "https://paper.id/test.jpg",
				MID:        sql.NullString{String: "0000", Valid: true},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				IndustryID: uuid.NewString(),
				DistrictId: 123,
				RiskLevel: sql.NullString{
					String: "INVALID_LEVEL",
					Valid:  true,
				},
			},
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = ""
				userRepo.On("FindUserByID", mock.Anything, mock.AnythingOfType("string")).Return(expectedUser, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: valid risk level LOW",
			input: &merchantModel.Merchant{
				UUID:       "merchant-id",
				Status:     constant.MerchantStatusActive,
				Name:       "test",
				Logo:       "https://paper.id/test.jpg",
				MID:        sql.NullString{String: "0000", Valid: true},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				IndustryID: uuid.NewString(),
				DistrictId: 123,
				RiskLevel: sql.NullString{
					String: constant.MerchantRiskLevelLow,
					Valid:  true,
				},
			},
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = ""
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*merchant.Merchant"),
				).Return(nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)

				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("0000", nil)
				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(nil)
				trxRepo.On(
					"CreateMerchantAuth",
					mock.Anything,
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(nil)

				mockString := "mockString"
				enc.On("GenerateRandomPKCS8Key").Return([]byte{}, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return(mockString)
				enc.On("Encrypt", mock.Anything, mock.Anything).Return(mockString, nil)

				rmq.On(
					"Publish",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				accountRepo.On(
					"Create",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: valid risk level HIGH",
			input: &merchantModel.Merchant{
				UUID:       "merchant-id",
				Status:     constant.MerchantStatusActive,
				Name:       "test",
				Logo:       "https://paper.id/test.jpg",
				MID:        sql.NullString{String: "0000", Valid: true},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				IndustryID: uuid.NewString(),
				DistrictId: 123,
				RiskLevel: sql.NullString{
					String: constant.MerchantRiskLevelHigh,
					Valid:  true,
				},
			},
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = ""
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*merchant.Merchant"),
				).Return(nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)

				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("0000", nil)
				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(nil)
				trxRepo.On(
					"CreateMerchantAuth",
					mock.Anything,
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(nil)

				mockString := "mockString"
				enc.On("GenerateRandomPKCS8Key").Return([]byte{}, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return(mockString)
				enc.On("Encrypt", mock.Anything, mock.Anything).Return(mockString, nil)

				rmq.On(
					"Publish",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				accountRepo.On(
					"Create",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: empty risk level (should be allowed)",
			input: &merchantModel.Merchant{
				UUID:       "merchant-id",
				Status:     constant.MerchantStatusActive,
				Name:       "test",
				Logo:       "https://paper.id/test.jpg",
				MID:        sql.NullString{String: "0000", Valid: true},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				IndustryID: uuid.NewString(),
				DistrictId: 123,
				RiskLevel: sql.NullString{
					String: "",
					Valid:  false,
				},
			},
			mocksSetup: func(trxRepo *repositoryMocks.IMerchantRepository, userRepo *mockUser.IUserRepository, rmq *mockRabbitMq.RabbitMQExt, enc *mockEncrypt.ICrypto, accountRepo *repositoryMocks.IAccountRepository, industrySvc *mocks.IIndustryService, countrySvc *mocks.ICountryService) {
				expectedUser.MerchantId = ""
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*merchant.Merchant"),
				).Return(nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)

				trxRepo.On(
					"GenerateNewMID",
					mock.Anything,
				).Return("0000", nil)
				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(nil)
				trxRepo.On(
					"CreateMerchantAuth",
					mock.Anything,
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(nil)

				mockString := "mockString"
				enc.On("GenerateRandomPKCS8Key").Return([]byte{}, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return(mockString)
				enc.On("Encrypt", mock.Anything, mock.Anything).Return(mockString, nil)

				rmq.On(
					"Publish",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				accountRepo.On(
					"Create",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			userRepo := mockUser.NewIUserRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			rabbitMqMock := mockRabbitMq.NewRabbitMQExt(t)
			jwtMock := mockJWT.NewIJwt(t)
			encryptMock := mockEncrypt.NewICrypto(t)
			accountRepoMock := repositoryMocks.NewIAccountRepository(t)
			accountSvc := mocks.NewIAccountService(t)
			industrySvc := mocks.NewIIndustryService(t)
			countrySvc := mocks.NewICountryService(t)

			tc.mocksSetup(merchantRepo, userRepo, rabbitMqMock, encryptMock, accountRepoMock, industrySvc, countrySvc)

			trxSvc := New(merchantRepo, loggerMock, userRepo, jwtMock, rabbitMqMock, encryptMock, WithAccountRepository(accountRepoMock), WithAccountService(accountSvc), WithLocationRepository(locRepo), WithIndustryService(industrySvc), WithCountryService(countrySvc), WithVaultTransit(vaultTransit))

			err := trxSvc.Create(context.Background(), tc.input, mock.Anything)
			if tc.wantErr {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tc.expectedError))
			} else {
				assert.NoError(t, err)
			}

			merchantRepo.AssertExpectations(t)
		})
	}
}
