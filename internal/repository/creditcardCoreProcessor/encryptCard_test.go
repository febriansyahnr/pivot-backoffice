package creditcardCoreProcessorRepository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/creditcardCoreProcessor"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestEncryptCardData(t *testing.T) {
	successResponse := `{
		"code": "200",
		"message": "success",
		"data": {
			"client_reference_id": "test-ref-123",
			"encrypted_card": "encrypted-card-data",
			"encrypted_card_informations": {
				"first_8_digits": "12345678",
				"first_6_digits": "123456",
				"last_4_digits": "7890",
				"expiry_month": "12",
				"expiry_year": "25",
				"has_associated_cvc": true,
				"fingerprint": "test-fingerprint"
			},
			"device_information": {
				"type": "web",
				"user_agent": "test-agent",
				"ip_address": "127.0.0.1"
			},
			"created_at": "2023-01-01T00:00:00Z"
		}
	}`

	testCases := []struct {
		name      string
		request   *creditcardCoreProcessorModel.EncryptCardRequest
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name: "SUCCESS: encrypt card data",
			request: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "test-merchant-id",
				ClientReferenceID: "test-ref-123",
				CardRequest: creditcardCoreProcessorModel.EncryptCardDetailRequest{
					Number:      "1234567890123456",
					ExpiryMonth: "12",
					ExpiryYear:  "25",
					CVC:         "123",
					NameOnCard:  "Test User",
				},
				DeviceInformation: creditcardCoreProcessorModel.DeviceInformation{
					Type:      "web",
					UserAgent: "test-agent",
					IpAddress: "127.0.0.1",
				},
			},
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name: "ERROR: HTTP error during request",
			request: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "test-merchant-id",
				ClientReferenceID: "test-ref-123",
			},
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "ERROR: invalid JSON response",
			request: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "test-merchant-id",
				ClientReferenceID: "test-ref-123",
			},
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{invalid json}`), 200, nil)
			},
		},
		{
			name: "ERROR: 400 client error",
			request: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "test-merchant-id",
				ClientReferenceID: "test-ref-123",
			},
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","message":"bad request"}`), 400, nil)
			},
		},
		{
			name: "ERROR: 500 server error",
			request: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "test-merchant-id",
				ClientReferenceID: "test-ref-123",
			},
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","message":"internal server error"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: "https://test-api.com"}}
			secret := &config.Secret{SnapCoreSecret: struct {
				InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
			}{InternalServiceKey: "test-key"}}

			repo := New(conf, secret, mockLogger, mockHttp)
			result, err := repo.EncryptCardData(context.Background(), tc.request)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "test-ref-123", result.ClientReferenceID)
				assert.Equal(t, "encrypted-card-data", result.EncryptedCard)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestGetEncryptedCardData(t *testing.T) {
	successResponse := `{
		"code": "200",
		"message": "success",
		"data": {
			"client_reference_id": "test-ref-123",
			"encrypted_card": "encrypted-card-data",
			"encrypted_card_informations": {
				"first_8_digits": "12345678",
				"first_6_digits": "123456",
				"last_4_digits": "7890",
				"expiry_month": "12",
				"expiry_year": "25",
				"has_associated_cvc": true,
				"fingerprint": "test-fingerprint"
			},
			"device_information": {
				"type": "web",
				"user_agent": "test-agent",
				"ip_address": "127.0.0.1"
			},
			"created_at": "2023-01-01T00:00:00Z"
		}
	}`

	testCases := []struct {
		name       string
		merchantId string
		cardId     string
		wantError  bool
		setupMock  func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:       "SUCCESS: get encrypted card data",
			merchantId: "test-merchant-id",
			cardId:     "test-card-id",
			wantError:  false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:       "ERROR: HTTP error during request",
			merchantId: "test-merchant-id",
			cardId:     "test-card-id",
			wantError:  true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:       "ERROR: invalid JSON response",
			merchantId: "test-merchant-id",
			cardId:     "test-card-id",
			wantError:  true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{invalid json}`), 200, nil)
			},
		},
		{
			name:       "ERROR: 404 not found",
			merchantId: "test-merchant-id",
			cardId:     "invalid-card-id",
			wantError:  true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"404","message":"card not found"}`), 404, nil)
			},
		},
		{
			name:       "ERROR: 500 server error",
			merchantId: "test-merchant-id",
			cardId:     "test-card-id",
			wantError:  true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","message":"internal server error"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: "https://test-api.com"}}
			secret := &config.Secret{SnapCoreSecret: struct {
				InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
			}{InternalServiceKey: "test-key"}}

			repo := New(conf, secret, mockLogger, mockHttp)
			result, err := repo.GetEncryptedCardData(context.Background(), tc.merchantId, tc.cardId)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "test-ref-123", result.ClientReferenceID)
				assert.Equal(t, "encrypted-card-data", result.EncryptedCard)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}