package creditcardCoreProcessorRepository_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/creditcardCoreProcessor"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInquiryTransaction(t *testing.T) {
	successResponse := `{
		"code": "200",
		"message": "success",
		"data": {
			"merchant_id": "merchant-123",
			"client_reference_id": "ref-123",
			"processor_reference_id": "proc-ref-123",
			"status": "SUCCESS",
			"amount": 100000
		}
	}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Inquiry transaction",
			wantError: false,
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
			name:      "ERROR: HTTP request error",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 0, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: unmarshal JSON response",
			wantError: true,
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
			name:      "ERROR: got 400 status code",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","message":"bad request"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 400 status code with error field",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"validation error"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 404 status code",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"404","message":"not found"}`), 404, nil)
			},
		},
		{
			name:      "ERROR: got 500 status code",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","message":"internal server error"}`), 500, nil)
			},
		},
		{
			name:      "ERROR: got 500 status code with error field",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"server error"}`), 500, nil)
			},
		},
		{
			name:      "ERROR: got 503 status code",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"503","message":"service unavailable"}`), 503, nil)
			},
		},
		{
			name:      "SUCCESS: with invalid data conversion (logs error but returns success)",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				// Response with data that can't be converted to expected struct
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"200","message":"success","data":"invalid_string_instead_of_object"}`), 200, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{
				CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
					BaseUrl: "https://creditcard-core.test",
				},
			}
			secret := &config.Secret{
				CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
					InternalServiceKey: "test-key",
				},
			}

			repo := New(conf, secret, mockLogger, mockHttp)
			payload := &creditcardModel.InquiryTransactionRequest{
				MerchantID:           "merchant-123",
				ClientReferenceID:    "ref-123",
				ProcessorReferenceID: "proc-ref-123",
			}

			result, err := repo.InquiryTransaction(context.Background(), payload)

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Note: result can be nil if data conversion fails (error is logged but not returned)
				if tc.name != "SUCCESS: with invalid data conversion (logs error but returns success)" {
					assert.NotNil(t, result)
				}
			}

			mockHttp.AssertExpectations(t)
		})
	}
}
