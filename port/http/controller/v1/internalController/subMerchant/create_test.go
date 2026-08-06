package submerchant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchantSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSubMerchant(t *testing.T) {

	parentID := uuid.Max
	response := &merchantModel.Merchant{}

	payloadRequest := &merchantModel.MerchantRequest{
		Name:              "test",
		MerchantStatus:    constant.MerchantStatusActive,
		ShortName:         "t",
		Description:       "test",
		Address:           "malang",
		DistrictId:        123,
		PostCode:          "123",
		Logo:              "test",
		MerchantEmail:     "test@test.com",
		MerchantPhone:     "test",
		BusinessType:      "test",
		BusinessStructure: "test",
		BusinessCountry:   "test",
		PICName:           "test",
		PICEmail:          "test@test.com",
		PICPhone:          "test",
		PICJobTitle:       "test",
		ParentID:          parentID.String(),
		UserID:            "2",
		BankAccount: &merchantModel.MerchantBankAccountRequest{
			ChannelCode: "BRI", AccountNumber: "123456",
		},
		DigitalStatus:   "Digital",
		CountryOfEntity: "ID",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant := &merchantModel.Merchant{
					UUID: "some-merchant-id",
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchant, nil)
				merchantSvcMocks.On("CreateSubMerchant", mock.Anything, mock.Anything).Return(response, nil)
			},
			expectedStatus: 200,
		},
		{
			name:        "ERROR: Error create sub merchant",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant := &merchantModel.Merchant{
					UUID: "some-merchant-id",
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchant, nil)
				merchantSvcMocks.On("CreateSubMerchant", mock.Anything, mock.Anything).
					Return(nil, errors.New("error"))
			},
			expectedStatus: 500,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant := &merchantModel.Merchant{
					UUID: "some-merchant-id",
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchant, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"client_transaction_id": "12345abcde"}`),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant := &merchantModel.Merchant{
					UUID: "some-merchant-id",
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchant, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "SUCCESS: Old merchant (before cutoff) - missing website fields bypassed",
			requestBody: []byte(`{
				"name": "Test Merchant",
				"shortName": "TM",
				"logo": "logo.png",
				"address": "Jakarta",
				"postCode": "12345",
				"merchantEmail": "test@test.com",
				"merchantPhone": "123456",
				"businessType": "retail",
				"businessStructure": "PT",
				"businessCountry": "ID",
				"picName": "John",
				"picEmail": "john@test.com",
				"picPhone": "654321",
				"bankAccount": {
					"channelCode": "BRI",
					"accountNumber": "123456"
				}
			}`),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Merchant created before cutoff date (2025-01-10)
				oldMerchant := &merchantModel.Merchant{
					UUID:      "old-merchant-id",
					CreatedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(oldMerchant, nil)
				merchantSvcMocks.On("CreateSubMerchant", mock.Anything, mock.Anything).Return(response, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "SUCCESS: New merchant (after cutoff) - without feature flag, no enforcement",
			requestBody: []byte(`{
				"name": "Test Merchant",
				"shortName": "TM",
				"logo": "logo.png",
				"address": "Jakarta",
				"postCode": "12345",
				"merchantEmail": "test@test.com",
				"merchantPhone": "123456",
				"businessType": "retail",
				"businessStructure": "PT",
				"businessCountry": "ID",
				"picName": "John",
				"picEmail": "john@test.com",
				"picPhone": "654321",
				"bankAccount": {
					"channelCode": "BRI",
					"accountNumber": "123456"
				}
			}`),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Merchant created before cutoff date (old merchant) - backward compatibility applies
				// Without feature flag, all mandatory fields (address + industry) are bypassed
				oldMerchant := &merchantModel.Merchant{
					UUID:      "old-merchant-id",
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(oldMerchant, nil)
				merchantSvcMocks.On("CreateSubMerchant", mock.Anything, mock.Anything).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "SUCCESS: New merchant (after cutoff) - all mandatory fields provided",
			requestBody: []byte(`{
				"name": "Test Merchant",
				"shortName": "TM",
				"logo": "logo.png",
				"website": "https://test.com",
				"address": "Jakarta",
				"districtId": 123,
				"postCode": "12345",
				"merchantEmail": "test@test.com",
				"merchantPhone": "123456",
				"businessType": "retail",
				"businessStructure": "PT",
				"businessCountry": "ID",
				"parentIndustry": "Retail",
				"childIndustry": "Fashion",
				"mcc": "5311",
				"countryOfEntity": "ID",
				"digitalStatus": "Digital",
				"picName": "John",
				"picEmail": "john@test.com",
				"picPhone": "654321",
				"bankAccount": {
					"channelCode": "BRI",
					"accountNumber": "123456"
				}
			}`),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Merchant created after cutoff date (2025-01-20) with all fields provided
				newMerchant := &merchantModel.Merchant{
					UUID:      "new-merchant-id",
					CreatedAt: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(newMerchant, nil)
				merchantSvcMocks.On("CreateSubMerchant", mock.Anything, mock.Anything).Return(response, nil)
			},
			expectedStatus: 200,
		},
		{
			name: "ERROR: Old merchant (before cutoff) - other required fields still enforced",
			requestBody: []byte(`{
				"shortName": "TM",
				"logo": "logo.png"
			}`),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Merchant created before cutoff date (2025-01-10)
				oldMerchant := &merchantModel.Merchant{
					UUID:      "old-merchant-id",
					CreatedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(oldMerchant, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "SUCCESS: Old merchant (before cutoff) - missing industry fields bypassed",
			requestBody: []byte(`{
				"name": "Test Merchant",
				"shortName": "TM",
				"logo": "logo.png",
				"address": "Jakarta",
				"postCode": "12345",
				"merchantEmail": "test@test.com",
				"merchantPhone": "123456",
				"businessType": "retail",
				"businessStructure": "PT",
				"businessCountry": "ID",
				"picName": "John",
				"picEmail": "john@test.com",
				"picPhone": "654321",
				"bankAccount": {
					"channelCode": "BRI",
					"accountNumber": "123456"
				}
			}`),
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				// Merchant created before cutoff date (2025-01-10) - industry fields (ParentIndustry, ChildIndustry, MCC, CountryOfEntity, DigitalStatus) are not provided
				oldMerchant := &merchantModel.Merchant{
					UUID:      "old-merchant-id",
					CreatedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(oldMerchant, nil)
				merchantSvcMocks.On("CreateSubMerchant", mock.Anything, mock.Anything).Return(response, nil)
			},
			expectedStatus: 200,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			merchantSvc := mockMerchantSvc.NewIMerchantService(t)
			accountSvc := mockMerchantSvc.NewIAccountService(t)
			orchestratorSvc := mockMerchantSvc.NewIOrchestratorService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(merchantSvc, mockRmq)

			mc := New(merchantSvc, accountSvc, orchestratorSvc, mockValidator)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/sub-merchants/create", bytes.NewBuffer(tt.requestBody))

			merchantAuth := &merchantModel.MerchantAuthTokenClaims{
				MerchantId: "aec6636d-7a02-4d93-a4c5-006b9c235068", // NOSONAR
			}
			ctx = context.WithValue(ctx, constant.CtxMerchantInfo, merchantAuth)
			ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, merchantAuth.MerchantId)

			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Create)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			merchantSvc.AssertExpectations(t)
		})
	}
}
