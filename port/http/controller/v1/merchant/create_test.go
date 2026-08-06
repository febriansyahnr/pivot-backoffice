package merchant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchant_Create(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID: uuid.NewString(),
		Role: constant.RoleAdmin,
	}

	payloadRequest := &merchantModel.MerchantRequest{
		Name:              "PT. Paper Indonesia",
		MerchantStatus:    constant.MerchantStatusActive,
		ShortName:         "Paper",
		Description:       "PT. Paper Indonesia is a company that provides a platform for people to buy and sell paper products.",
		Website:           "https://paper.co.id",
		Address:           "Jakarta",
		DistrictId:        3885,
		PostCode:          "65146",
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
		ParentIndustry:    "Technology",
		ChildIndustry:     "Fintech",
		MCC:               "5734",
		DigitalStatus:     "Digital",
		CountryOfEntity:   "ID",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
		userClaim      *user.UserTokenClaims
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
			name:           "ERROR: User already have a merchant",
			requestBody:    payloadRequestByte,
			mockSetup:      func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			expectedStatus: http.StatusForbidden,
			userClaim: &user.UserTokenClaims{
				UUID:       uuid.NewString(),
				MerchantId: uuid.NewString(),
				Role:       constant.RoleAdmin,
			},
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
			name:           "ERROR: User not in Context",
			mockSetup:      func(merchantSvcMocks *mockMerchant.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {},
			userClaim:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(mockMerchantSvc, mockRmq)

			mc := New(mockMerchantSvc, mockValidator, mockRmq, WithConfig(&config.Config{
				WithdrawalConfig: config.WithdrawalConfig{
					AutoWithdrawalDefaultState: constant.AutoWithdrawalStateOFF,
				},
			}))

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
