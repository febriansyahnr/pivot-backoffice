package subMerchant

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSubMerchant(t *testing.T) {

	parentID := uuid.Max
	response := &merchantModel.Merchant{}

	payloadRequest := &merchantModel.CreateSubMerchantRequest{
		Name:              "test",
		MerchantStatus:    constant.MerchantStatusActive,
		ShortName:         "t",
		Description:       "test",
		Website:           "https://test.com",
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
		ParentIndustry:    "test",
		ChildIndustry:     "test",
		MCC:               "test",
		BankAccount: &merchantModel.MerchantBankAccountRequest{
			ChannelCode: "BRI", AccountNumber: "123456",
		},
		KYCStatus:       constant.MerchantKYCTypeNonKYC,
		DigitalStatus:   "Digital",
		CountryOfEntity: "ID",
	}
	payloadRequestByte, err := json.Marshal(payloadRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name           string
		requestBody    []byte
		mockSetup      func(merchantSvcMocks *mockSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		expectedStatus int
	}{
		{
			name:        "SUCCESS",
			requestBody: payloadRequestByte,
			mockSetup: func(merchantSvcMocks *mockSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
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
			mockSetup: func(merchantSvcMocks *mockSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant := &merchantModel.Merchant{
					UUID: "some-merchant-id",
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchant, nil)
				merchantSvcMocks.On("CreateSubMerchant", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
			expectedStatus: 500,
		},
		{
			name:        "ERROR: when the payload was invalid, then should return error",
			requestBody: []byte("invalid-payload"),
			mockSetup: func(merchantSvcMocks *mockSvc.IMerchantService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				merchant := &merchantModel.Merchant{
					UUID: "some-merchant-id",
				}
				merchantSvcMocks.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(merchant, nil)
			},
			expectedStatus: 400,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockMerchantSvc := mockSvc.NewIMerchantService(t)
			mockAccountSvc := mockSvc.NewIAccountService(t)
			mockOrchestratorSvc := mockSvc.NewIOrchestratorService(t)
			mockForbiddenSvc := mockSvc.NewIMerchantForbiddenUseCaseService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			tt.mockSetup(mockMerchantSvc, mockRmq)

			mc := New(mockMerchantSvc, mockAccountSvc, mockOrchestratorSvc, mockForbiddenSvc, mockValidator, mockRmq)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/api/v1/sub-merchants/create", bytes.NewBuffer(tt.requestBody))
			merchantID := uuid.NewString()
			ctx = context.WithValue(ctx, constant.CtxUserInfoKey, &user.UserTokenClaims{
				UUID:       uuid.NewString(),
				MerchantId: merchantID,
			})
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
			mockMerchantSvc.AssertExpectations(t)
		})
	}

}

func TestSetDefaultMerchantValuesCreate(t *testing.T) {
	tests := []struct {
		name     string
		payload  *merchantModel.CreateSubMerchantRequest
		merchant *merchantModel.Merchant
		expected *merchantModel.CreateSubMerchantRequest
	}{
		{
			name:    "All fields empty",
			payload: &merchantModel.CreateSubMerchantRequest{},
			merchant: &merchantModel.Merchant{
				Logo:          "default-logo",
				Website:       "https://default.com",
				MerchantEmail: "default-email@test.com",
				MerchantPhone: "123456789",
				PICJobTitle: sql.NullString{
					String: "Manager",
					Valid:  true,
				},
				PICPhone: "987654321",
				Address:  "default-address",
				PostCode: "12345",
			},
			expected: &merchantModel.CreateSubMerchantRequest{
				Logo:           "default-logo",
				MerchantStatus: constant.MerchantStatusActive,
				Website:        "https://default.com",
				MerchantEmail:  "default-email@test.com",
				MerchantPhone:  "123456789",
				PICJobTitle:    "Manager",
				PICPhone:       "987654321",
				Address:        "default-address",
				PostCode:       "12345",
			},
		},
		{
			name: "Some fields empty",
			payload: &merchantModel.CreateSubMerchantRequest{
				Logo:          "custom-logo",
				MerchantEmail: "custom-email@test.com",
			},
			merchant: &merchantModel.Merchant{
				Logo:          "default-logo",
				Website:       "https://default.com",
				MerchantEmail: "default-email@test.com",
				MerchantPhone: "123456789",
				PICJobTitle: sql.NullString{
					String: "Manager",
					Valid:  true,
				},
				PICPhone: "987654321",
				Address:  "default-address",
				PostCode: "12345",
			},
			expected: &merchantModel.CreateSubMerchantRequest{
				Logo:           "custom-logo",
				MerchantStatus: constant.MerchantStatusActive,
				Website:        "https://default.com",
				MerchantEmail:  "custom-email@test.com",
				MerchantPhone:  "123456789",
				PICJobTitle:    "Manager",
				PICPhone:       "987654321",
				Address:        "default-address",
				PostCode:       "12345",
			},
		},
		{
			name: "address and postal code empty and also the merchant, then should return empty",
			payload: &merchantModel.CreateSubMerchantRequest{
				Logo:          "custom-logo",
				MerchantEmail: "custom-email@test.com",
			},
			merchant: &merchantModel.Merchant{
				Logo:          "default-logo",
				Website:       "https://default.com",
				MerchantEmail: "default-email@test.com",
				MerchantPhone: "123456789",
				PICJobTitle: sql.NullString{
					String: "Manager",
					Valid:  true,
				},
				PICPhone: "987654321",
			},
			expected: &merchantModel.CreateSubMerchantRequest{
				Logo:           "custom-logo",
				MerchantStatus: constant.MerchantStatusActive,
				Website:        "https://default.com",
				MerchantEmail:  "custom-email@test.com",
				MerchantPhone:  "123456789",
				PICJobTitle:    "Manager",
				PICPhone:       "987654321",
				Address:        "-",
				PostCode:       "0",
			},
		},
		{
			name: "All fields filled",
			payload: &merchantModel.CreateSubMerchantRequest{
				Logo:           "custom-logo",
				MerchantStatus: constant.MerchantStatusInactive,
				Website:        "https://custom.com",
				MerchantEmail:  "custom-email@test.com",
				MerchantPhone:  "111111111",
				PICJobTitle:    "Director",
				PICPhone:       "222222222",
				Address:        "custom-address",
				PostCode:       "54321",
			},
			merchant: &merchantModel.Merchant{
				Logo:          "default-logo",
				Website:       "https://default.com",
				MerchantEmail: "default-email@test.com",
				MerchantPhone: "123456789",
				PICJobTitle: sql.NullString{
					String: "Manager",
					Valid:  true,
				},
				PICPhone: "987654321",
				Address:  "default-address",
				PostCode: "12345",
			},
			expected: &merchantModel.CreateSubMerchantRequest{
				Logo:           "custom-logo",
				MerchantStatus: constant.MerchantStatusInactive,
				Website:        "https://custom.com",
				MerchantEmail:  "custom-email@test.com",
				MerchantPhone:  "111111111",
				PICJobTitle:    "Director",
				PICPhone:       "222222222",
				Address:        "custom-address",
				PostCode:       "54321",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultMerchantValuesCreate(tt.payload, tt.merchant)
			assert.Equal(t, tt.expected, tt.payload)
		})
	}
}
