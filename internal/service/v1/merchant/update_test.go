package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/location"
	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	countryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/country"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantService_Update(t *testing.T) {
	keyStr := "key123"

	redisClient, redisClientMock := redismock.NewClientMock()

	testCases := []struct {
		name      string
		merchant  *merchant.UpdateMerchantRequest
		mockSetup func(
			mockRepo *mocks.IMerchantRepository,
			mockLocationRepo *mocks.IAddrLocationRepository,
			mockRabbitMq *rabbitMqMocks.RabbitMQExt,
			mockCrypto *mockEncrypt.ICrypto,

			redisMock redismock.ClientMock,
			countryService *mockService.ICountryService,
		)
		wantErr bool
	}{
		{
			name: "Success update merchant",
			merchant: &merchant.UpdateMerchantRequest{
				ID:              uuid.New().String(),
				DistrictId:      1,
				CountryOfEntity: "ID",
				KYMNotes:        "OPS Notes",
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,

				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{MID: sql.NullString{String: "123", Valid: true}, DistrictId: 1, CallbackApiKey: &keyStr}, nil)
				mockLocationRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(nil)
				mockRabbitMq.On(
					"Publish",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.PtrStringMockType(),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				redisMock.ExpectDel("key-1").SetVal(1)

				countryService.On("FindByCode", mock.Anything, mock.Anything).Return(&countryModel.Country{Code: "ID"}, nil)
			},
			wantErr: false,
		},
		{
			name: "Error get merchant id",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         uuid.New().String(),
				DistrictId: 1,
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,

				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, errors.New("error when find merchant by id"))
			},
			wantErr: true,
		},
		{
			name: "Error merchant not found",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         uuid.New().String(),
				DistrictId: 1,
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,

				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "Error getting district",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         "merchant-id",
				DistrictId: 2,
				Name:       "test",
				Logo:       "https://paper.id/test.jpg",
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,

				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{MID: sql.NullString{String: "123", Valid: true}, CallbackApiKey: &keyStr}, nil)
				mockLocationRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).
					Once().
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "District not found",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         uuid.New().String(),
				DistrictId: 1,
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,

				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				mockLocationRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "Error updating merchant KYM Notes",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         uuid.New().String(),
				DistrictId: 1,
				KYMNotes:   "Notes",
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,

				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{MID: sql.NullString{String: "123", Valid: true}, CallbackApiKey: &keyStr, Metadata: types.NullJSONText{Valid: true, JSONText: []byte(`{}}{}`)}}, nil)
				mockLocationRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)
			},
			wantErr: true,
		},
		{
			name: "Error updating merchant",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         uuid.New().String(),
				DistrictId: 1,
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,

				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{MID: sql.NullString{String: "123", Valid: true}, CallbackApiKey: &keyStr}, nil)
				mockLocationRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(errors.New("update error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: invalid risk level",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         uuid.New().String(),
				DistrictId: 1,
				RiskLevel:  "INVALID_LEVEL",
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,
				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{MID: sql.NullString{String: "123", Valid: true}, CallbackApiKey: &keyStr}, nil)
				mockLocationRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: valid risk level LOW",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         uuid.New().String(),
				DistrictId: 1,
				RiskLevel:  constant.MerchantRiskLevelLow,
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,
				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{MID: sql.NullString{String: "123", Valid: true}, CallbackApiKey: &keyStr}, nil)
				mockLocationRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(nil)
				mockRabbitMq.On("Publish", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), constant.PtrStringMockType(), mock.AnythingOfType("[]uint8")).Return(nil)
				redisMock.ExpectDel("key-1").SetVal(1)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: valid risk level MID_HIGH",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         uuid.New().String(),
				DistrictId: 1,
				RiskLevel:  constant.MerchantRiskLevelMidHigh,
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,
				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{MID: sql.NullString{String: "123", Valid: true}, CallbackApiKey: &keyStr}, nil)
				mockLocationRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(nil)
				mockRabbitMq.On("Publish", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), constant.PtrStringMockType(), mock.AnythingOfType("[]uint8")).Return(nil)
				redisMock.ExpectDel("key-1").SetVal(1)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: empty risk level preserves existing",
			merchant: &merchant.UpdateMerchantRequest{
				ID:         uuid.New().String(),
				DistrictId: 1,
				RiskLevel:  "",
			},
			mockSetup: func(mockRepo *mocks.IMerchantRepository,
				mockLocationRepo *mocks.IAddrLocationRepository,
				mockRabbitMq *rabbitMqMocks.RabbitMQExt,
				mockCrypto *mockEncrypt.ICrypto,
				redisMock redismock.ClientMock,
				countryService *mockService.ICountryService,
			) {
				existingMerchant := &merchant.Merchant{
					MID:            sql.NullString{String: "123", Valid: true},
					CallbackApiKey: &keyStr,
					RiskLevel:      sql.NullString{String: constant.MerchantRiskLevelHigh, Valid: true},
				}
				mockRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(existingMerchant, nil)
				mockLocationRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(nil)
				mockRabbitMq.On("Publish", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), constant.PtrStringMockType(), mock.AnythingOfType("[]uint8")).Return(nil)
				redisMock.ExpectDel("key-1").SetVal(1)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewIMerchantRepository(t)
			mockLocationRepo := mocks.NewIAddrLocationRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRabbitMq := rabbitMqMocks.NewRabbitMQExt(t)
			encryptMock := mockEncrypt.NewICrypto(t)
			accountService := mockService.NewIAccountService(t)
			countryService := mockService.NewICountryService(t)

			tc.mockSetup(mockRepo, mockLocationRepo, mockRabbitMq, encryptMock, redisClientMock, countryService)

			svc := New(mockRepo, mockLogger, nil, nil, mockRabbitMq, encryptMock,
				WithAccountService(accountService),
				WithLocationRepository(mockLocationRepo),
				WithRedisClient(redisExt.WrapRedisClient(redisClient, nil)),
				WithCountryService(countryService),
			)

			_, err := svc.Update(context.Background(), tc.merchant)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockLocationRepo.AssertExpectations(t)

			mockRabbitMq.AssertExpectations(t)
		})
	}
}
