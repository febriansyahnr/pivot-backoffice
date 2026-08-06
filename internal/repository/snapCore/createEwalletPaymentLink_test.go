package snapCoreRepository

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/ewallet"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSnapCoreRepositoryCreateEWalletPaymentLink(t *testing.T) {
	successResponse := `{
		"code": "200",
		"message": "Success",
		"data": {
			"responseCode": "2001600",
			"responseMessage": "Successful",
			"partnerReferenceNo": "20240101120000001",
			"referenceNo": "TXN0001",
			"webRedirectUrl": "https://redirect.example.com/payment",
			"apRedirectUrl": "shopeelink://payment/redirect"
		}
	}`

	testCases := []struct {
		name      string
		wantError bool
		errType   string
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: create ewallet payment link",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP request error",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(""), 0, errors.New("network error"))
			},
		},
		{
			name:      "ERROR: invalid JSON response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{invalid json}`), 200, nil)
			},
		},
		{
			name:      "ERROR: 400 bad request",
			wantError: true,
			errType:   httpResponse.HttpErrRequest,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","message":"Bad Request"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: 400 with error field",
			wantError: true,
			errType:   httpResponse.HttpErrRequest,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":{"amount":"invalid amount"}}`), 400, nil)
			},
		},
		{
			name:      "ERROR: 500 internal server error",
			wantError: true,
			errType:   httpResponse.HttpErrThirdParty,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","message":"Internal Server Error"}`), 500, nil)
			},
		},
		{
			name:      "ERROR: 500 with error field",
			wantError: true,
			errType:   httpResponse.HttpErrThirdParty,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":{"internal":"server error"}}`), 500, nil)
			},
		},
		{
			name:      "ERROR: 429 too many requests",
			wantError: true,
			errType:   httpResponse.HttpErrRequestLimitExceeded,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"429","message":"Too many requests"}`), http.StatusTooManyRequests, nil)
			},
		},
		{
			name:      "ERROR: 408 request timeout",
			wantError: true,
			errType:   httpResponse.HttpErrRequestTimeout,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"408","message":"Request timeout"}`), http.StatusRequestTimeout, nil)
			},
		},
		{
			name:      "ERROR: 502 bad gateway",
			wantError: true,
			errType:   httpResponse.HttpErrBadGateway,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"502","message":"Bad gateway"}`), http.StatusBadGateway, nil)
			},
		},
		{
			name:      "ERROR: 503 service unavailable",
			wantError: true,
			errType:   httpResponse.HttpErrServiceUnavailable,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"503","message":"Service unavailable"}`), http.StatusServiceUnavailable, nil)
			},
		},
		{
			name:      "ERROR: 504 gateway timeout",
			wantError: true,
			errType:   httpResponse.HttpErrRequestTimeout,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"504","message":"Gateway timeout"}`), http.StatusGatewayTimeout, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{SnapCoreConfig: config.SnapCoreConfig{BaseUrl: "https://api.example.com"}}
			secret := &config.Secret{SnapCoreSecret: struct {
				InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
			}{InternalServiceKey: "test-key"}}

			repo := New(conf, secret, mockLogger, mockHttp)

			request := &ewallet.EwalletPaymentRequest{
				OriginalReferenceId: "20240101120000001",
				Acquirer:            "shopee",
				Amount: commonModel.Amount{
					Value:    "10000",
					Currency: "IDR",
				},
				ValidUpTo: "2024-12-31T23:59:59+07:00",
				AdditionalInfo: map[string]interface{}{
					"merchantId": "MERCHANT001",
				},
			}

			result, err := repo.CreateEWalletPaymentLink(context.Background(), request)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.errType != "" {
					extractedErrType, _ := pkgErr.ExtractError(err)
					assert.Equal(t, tc.errType, extractedErrType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "2001600", result.ResponseCode)
				assert.Equal(t, "Successful", result.ResponseMessage)
			}

			mockHttp.AssertExpectations(t)
		})
	}
}

func TestSnapCoreRepositoryInquiryStatusEWalletPayment(t *testing.T) {
	successResponse := `{
			"responseCode": "2001600",
			"responseMessage": "Successful",
			"latestTransactionStatus": "01",
			"originalPartnerReferenceNo": "20240101120000001",
			"originalReferenceNo": "TXN0001"
		}`

	testCases := []struct {
		name      string
		wantError bool
		errType   string
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: inquiry ewallet payment",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP request error",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(""), 0, errors.New("network error"))
			},
		},
		{
			name:      "ERROR: invalid JSON response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{invalid json}`), 200, nil)
			},
		},
		{
			name:      "ERROR: 400 bad request",
			wantError: true,
			errType:   httpResponse.HttpErrRequest,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","message":"Bad Request"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: 400 with error field",
			wantError: true,
			errType:   httpResponse.HttpErrRequest,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":{"amount":"invalid amount"}}`), 400, nil)
			},
		},
		{
			name:      "ERROR: 500 internal server error",
			wantError: true,
			errType:   httpResponse.HttpErrThirdParty,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","message":"Internal Server Error"}`), 500, nil)
			},
		},
		{
			name:      "ERROR: 500 with error field",
			wantError: true,
			errType:   httpResponse.HttpErrThirdParty,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":{"internal":"server error"}}`), 500, nil)
			},
		},
		{
			name:      "ERROR: 429 too many requests",
			wantError: true,
			errType:   httpResponse.HttpErrRequestLimitExceeded,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"429","message":"Too many requests"}`), http.StatusTooManyRequests, nil)
			},
		},
		{
			name:      "ERROR: 408 request timeout",
			wantError: true,
			errType:   httpResponse.HttpErrRequestTimeout,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"408","message":"Request timeout"}`), http.StatusRequestTimeout, nil)
			},
		},
		{
			name:      "ERROR: 502 bad gateway",
			wantError: true,
			errType:   httpResponse.HttpErrBadGateway,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"502","message":"Bad gateway"}`), http.StatusBadGateway, nil)
			},
		},
		{
			name:      "ERROR: 503 service unavailable",
			wantError: true,
			errType:   httpResponse.HttpErrServiceUnavailable,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"503","message":"Service unavailable"}`), http.StatusServiceUnavailable, nil)
			},
		},
		{
			name:      "ERROR: 504 gateway timeout",
			wantError: true,
			errType:   httpResponse.HttpErrRequestTimeout,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"504","message":"Gateway timeout"}`), http.StatusGatewayTimeout, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{SnapCoreConfig: config.SnapCoreConfig{BaseUrl: "https://api.example.com"}}
			secret := &config.Secret{SnapCoreSecret: struct {
				InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
			}{InternalServiceKey: "test-key"}}

			repo := New(conf, secret, mockLogger, mockHttp)

			request := &ewallet.EWalletInquiryStatusRequest{
				TransactionID: uuid.NewString(),
			}

			result, err := repo.InquiryStatusEWalletPayment(context.Background(), request)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.errType != "" {
					extractedErrType, _ := pkgErr.ExtractError(err)
					assert.Equal(t, tc.errType, extractedErrType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "2001600", result.ResponseCode)
				assert.Equal(t, "Successful", result.ResponseMessage)
			}

			mockHttp.AssertExpectations(t)
		})
	}
}
