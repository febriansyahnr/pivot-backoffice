package merchant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/test"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantCreate(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	payloadRequest := &merchantModel.CRMCreateMerchantRequest{
		Name:              "PT. Paper Indonesia",
		ShortName:         "Paper",
		Description:       "PT. Paper Indonesia is a company that provides a platform for people to buy and sell paper products.",
		Website:           "https://paper.co.id",
		Address:           "Jakarta",
		DistrictId:        123,
		PostCode:          "123",
		Logo:              "https://paper.co.id/logo.png",
		MerchantEmail:     "test@gmail.com",
		MerchantPhone:     "081234567890",
		PICEmail:          "testing@gmail.com",
		PICPhone:          "081234567890",
		BusinessType:      "type1",
		BusinessStructure: "pt",
		BusinessCountry:   "IDN",
		PICName:           "pic",
		PICJobTitle:       "tester",
		MerchantStatus:    constant.MerchantStatusActive,
		AutoWithdrawal:    util.ValueToPtr("OFF"),
		ParentIndustry:    "Technology",
		ChildIndustry:     "Fintech",
		MCC:               "5734",
		DigitalStatus:     "Digital",
		CountryOfEntity:   "ID",
	}

	payloadRequestNonKYC := &merchantModel.CRMCreateMerchantRequest{
		Name:              "PT. Paper Indonesia",
		ShortName:         "Paper",
		Description:       "PT. Paper Indonesia is a company that provides a platform for people to buy and sell paper products.",
		Website:           "https://paper.co.id",
		Address:           "Jakarta",
		DistrictId:        123,
		PostCode:          "123",
		Logo:              "https://paper.co.id/logo.png",
		MerchantEmail:     "test@gmail.com",
		MerchantPhone:     "081234567890",
		PICEmail:          "testing@gmail.com",
		PICPhone:          "081234567890",
		BusinessType:      "type1",
		BusinessStructure: "pt",
		BusinessCountry:   "IDN",
		PICName:           "pic",
		PICJobTitle:       "tester",
		MerchantStatus:    constant.MerchantStatusActive,
		KYCStatus:         constant.MerchantKYCTypeNonKYC,
		AutoWithdrawal:    util.ValueToPtr("OFF"),
		ParentIndustry:    "Technology",
		ChildIndustry:     "Fintech",
		MCC:               "5734",
		DigitalStatus:     "Digital",
		CountryOfEntity:   "ID",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	payloadRequestNonKYCByte, err := json.Marshal(payloadRequestNonKYC)
	assert.NoError(t, err)

	testCases := []struct {
		name                  string
		autoWithdrawalEnabled bool
		requestBody           []byte
		mockSetup             func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus        int
		userClaim             *user.UserTokenClaims
		withConfig            *config.Config
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*merchant.Merchant"),
						mock.Anything).
					Return(nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: 200,
			userClaim:      validUserClaims,
		},
		{
			name:                  "SUCCESS with auto withdrawal enabled and enable all payment method",
			autoWithdrawalEnabled: true,
			requestBody:           payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*merchant.Merchant"),
						mock.Anything).
					Return(nil)
				merchantSvcMocks.
					On("EnableAllPaymentMethod", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).
					Return(constant.ErrSomeErrorForUnitTest).Once()

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			withConfig: &config.Config{
				Environment: constant.EnvironmentStaging,
			},
			expectedStatus: 200,
			userClaim:      validUserClaims,
		},
		{
			name:                  "SUCCESS non KYC with auto withdrawal enabled",
			autoWithdrawalEnabled: true,
			requestBody:           payloadRequestNonKYCByte,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*merchant.Merchant"),
						mock.Anything).
					Return(nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: 200,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"client_transaction_id": "12345abcde"}`),
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Bad Request - Missing Required districtId",
			requestBody: []byte(`{"name":"Test","merchantStatus":"ACTIVE","shortName":"TEST","address":"Test Address","postcode":"12345","logo":"https://test.com/logo.png","merchantEmail":"test@test.com","merchantPhone":"081234567890","businessType":"type1","businessStructure":"pt","businessCountry":"IDN","picName":"pic","picEmail":"pic@test.com","picPhone":"081234567890"}`),
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:        "ERROR: Service Error",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*merchant.Merchant"),
						mock.Anything).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name:        "SUCCESS - verify merchant fields are mapped correctly",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchantSvcMocks.
					On(
						"Create",
						mock.Anything,
						mock.AnythingOfType("*merchant.Merchant"),
						mock.Anything).
					Run(func(args mock.Arguments) {
						merchant := args.Get(1).(*merchantModel.Merchant)
						assert.True(t, merchant.BusinessType.Valid)
						assert.Equal(t, "type1", merchant.BusinessType.String)
						assert.True(t, merchant.BusinessStructure.Valid)
						assert.Equal(t, "pt", merchant.BusinessStructure.String)
						assert.True(t, merchant.BusinessCountry.Valid)
						assert.Equal(t, "IDN", merchant.BusinessCountry.String)
						assert.True(t, merchant.PICName.Valid)
						assert.Equal(t, "pic", merchant.PICName.String)
						assert.True(t, merchant.PICJobTitle.Valid)
						assert.Equal(t, "tester", merchant.PICJobTitle.String)
					}).
					Return(nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: 200,
			userClaim:      validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockUserSvc := mockUser.NewIUserService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			cfg := &config.Config{
				WithdrawalConfig: config.WithdrawalConfig{
					AutoWithdrawalDefaultState: constant.AutoWithdrawalStateOFF,
				},
			}
			_, pdkLogger, err := test.SetupLogger()

			assert.NoError(t, err)
			tt.mockSetup(mockMerchantSvc, mockRmq)

			// replace the default config with the one from the test case
			if tt.withConfig != nil {
				cfg = tt.withConfig
			}

			if tt.autoWithdrawalEnabled {
				cfg.WithdrawalConfig.AutoWithdrawalDefaultState = constant.AutoWithdrawalStateON
			}

			mc := New(mockMerchantSvc, mockUserSvc, mockValidator, mockRmq, WithConfig(cfg), WithLogger(pdkLogger))

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/merchants/create", bytes.NewBuffer(tt.requestBody))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Create)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}
