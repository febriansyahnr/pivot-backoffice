package creditcardCoreProcessorRepository_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/creditcardCoreProcessor"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	encryptionMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionList(t *testing.T) {
	successResponse := `{"example": "ok"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Get List",
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
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)

			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarsal}`), 200, nil)

			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
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
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)

			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := logger.NewZapLogger(logger.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{SnapCoreSecret: struct {
				InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
			}{InternalServiceKey: ""}}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.GetTransactionList(context.Background(), &creditcardCoreProcessorModel.GetTransactionListRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)

		})
	}
}

func TestGetBinDetailByBinNumber(t *testing.T) {
	log := logMock.NewILogger(t)
	httpClient := httpMocks.NewIHTTPRequest(t)

	repo := New(&config.Config{}, &config.Secret{}, log, httpClient)

	binNumber := "123456" // NOSONAR
	merchantId := "038109b5-6b23-42c2-86c5-78b8824ab346"
	logMessage := "BIN details request for merchant ID " + merchantId

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *creditcardCoreProcessorModel.GetBinDetailResponse
	}{
		{
			name: "ERROR:Send HTTP client",
			setupMock: func() {
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, 0, assert.AnError)

				log.On("Error", mock.Anything, "Failed while send request get bin detail by bin number", mock.Anything).Once().Return()
				log.On(
					"Info", mock.Anything, logMessage, logger.String("binNumber", binNumber), logger.Any("response", map[string]any{
						"responseBody": string([]byte{}),
						"statusCode":   int(0),
					}), mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrors.New(httpResponse.HttpErrInternal, assert.AnError),
		},
		{
			name: "ERROR:BIN detail not found",
			setupMock: func() {
				response := []byte(`{"code":"44","error":"NOT_FOUND","error_type":"API_ERROR"}`)

				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).Once().Return(response, http.StatusNotFound, nil)
				log.On(
					"Info", mock.Anything, logMessage, logger.String("binNumber", binNumber), logger.Any("response", map[string]any{
						"responseBody": string(response),
						"statusCode":   http.StatusNotFound,
					}), mock.Anything,
				).Once().Return()
			},
		},
		{
			name: "ERROR:Unmarshal response",
			setupMock: func() {
				response := []byte(`Bad Gateway`)

				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).Once().Return(response, http.StatusBadGateway, nil)

				log.On("Error", mock.Anything, "Failed while unmarshalling bin response details", mock.Anything).Once().Return()
				log.On(
					"Info", mock.Anything, logMessage, logger.String("binNumber", binNumber), logger.Any("response", map[string]any{
						"responseBody": string(response),
						"statusCode":   http.StatusBadGateway,
					}), mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrors.New(httpResponse.HttpErrInternal, errors.New("invalid character 'B' looking for beginning of value")),
		},
		{
			name: "ERROR:Internal server error",
			setupMock: func() {
				response := []byte(`{"code":"99","error":"INTERNAL_SERVER_ERROR","error_type":"API_ERROR"}`)

				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).Once().Return(response, http.StatusInternalServerError, nil)

				log.On(
					"Info", mock.Anything, logMessage, logger.String("binNumber", binNumber), logger.Any("response", map[string]any{
						"responseBody": string(response),
						"statusCode":   http.StatusInternalServerError,
					}), mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrors.New(httpResponse.HttpErrInternal, errors.New("Internal Server Error")),
		},
		{
			name: "ERROR:Bad request",
			setupMock: func() {
				response := []byte(`{"code":"40","error":"API_VALIDATION_ERROR","error_type":"API_ERROR"}`)

				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).Once().Return(response, http.StatusBadRequest, nil)

				log.On(
					"Info", mock.Anything, logMessage, logger.String("binNumber", binNumber), logger.Any("response", map[string]any{
						"responseBody": string(response),
						"statusCode":   http.StatusBadRequest,
					}), mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrors.New(httpResponse.HttpErrRequest, errors.New("Api Validation Error")),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				response := []byte(`{"code":"00","data":{"bin_number":"123456","card_type":"DEBIT","card_brand":"MASTERCARD","card_level":"CLASSIC","issuer_name":"BCA","issuer_country":"ID","currency":"IDR"}}`)

				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).Once().Return(response, http.StatusOK, nil)

				log.On(
					"Info", mock.Anything, logMessage, logger.String("binNumber", binNumber), logger.Any("response", map[string]any{
						"responseBody": string(response),
						"statusCode":   http.StatusOK,
					}), mock.Anything,
				).Once().Return()
			},
			wantResult: &creditcardCoreProcessorModel.GetBinDetailResponse{
				BinNumber:     binNumber,
				CardType:      "DEBIT",
				CardBrand:     "MASTERCARD",
				CardLevel:     "CLASSIC",
				IssuerName:    "BCA",
				IssuerCountry: "ID",
				Currency:      "IDR",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetBinDetailByBinNumber(t.Context(), merchantId, binNumber)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			log.AssertExpectations(t)
			httpClient.AssertExpectations(t)
		})
	}
}

func TestGetCardEncryptionPublicKey(t *testing.T) {
	const (
		// validKEK: base64 of 32 zero bytes
		validKEK = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
		// validIV: base64 of 16 zero bytes
		validIV = "AAAAAAAAAAAAAAAAAAAAAA=="
		// validDataJSON: "QUE9PQ==" → stdDecode → "AA==" → urlDecode → []byte{0x00}
		validDataJSON = `{"data":"QUE9PQ=="}`
	)

	validSecret := config.CreditcardCoreProcessorSecret{
		EncryptionPublicKeySecret: validKEK,
		EncryptionPublicKeyIV:     validIV,
	}

	logger := logMock.NewILogger(t)
	httpClient := httpMocks.NewIHTTPRequest(t)
	crypto := encryptionMocks.NewCryptoProvider(t)

	secret := &config.Secret{}
	conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
	repo := New(conf, secret, logger, httpClient, WithCryptoProvider(crypto))

	tests := []struct {
		name            string
		secret          config.CreditcardCoreProcessorSecret
		setupMock       func()
		wantErrContains string
	}{
		{
			name:   "ERROR: HTTP GET fails",
			secret: validSecret,
			setupMock: func() {
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(nil, 0, assert.AnError)
				logger.On("Error", mock.Anything, "Failed while get encryption public key", mock.Anything).Once().Return()
			},
			wantErrContains: assert.AnError.Error(),
		},
		{
			name:   "ERROR: HTTP status 4xx",
			secret: validSecret,
			setupMock: func() {
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(`bad request`), http.StatusBadRequest, nil)
				logger.On("Error", mock.Anything, "Failed to get encryption public key, received http code 400", mock.Anything).Once().Return()
			},
			wantErrContains: "bad request",
		},
		{
			name:   "ERROR: HTTP status 5xx",
			secret: validSecret,
			setupMock: func() {
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(`internal error`), http.StatusInternalServerError, nil)
				logger.On("Error", mock.Anything, "Failed to get encryption public key, received http code 500", mock.Anything).Once().Return()
			},
			wantErrContains: "internal error",
		},
		{
			name:   "ERROR: unmarshal response body fails",
			secret: validSecret,
			setupMock: func() {
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(`{invalid json}`), http.StatusOK, nil)
				logger.On("Error", mock.Anything, "Failed to unmarshal response body", mock.Anything).Once().Return()
			},
			wantErrContains: "invalid character 'i' looking for beginning of object key string",
		},
		{
			name: "ERROR: invalid EncryptionPublicKeySecret",
			secret: config.CreditcardCoreProcessorSecret{
				EncryptionPublicKeySecret: "!!!not-valid-base64!!!",
				EncryptionPublicKeyIV:     validIV,
			},
			setupMock: func() {
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(validDataJSON), http.StatusOK, nil)
				logger.On("Error", mock.Anything, "Failed to decrypt card encryption public key", mock.Anything).Once().Return()
			},
			wantErrContains: "decode kek: illegal base64 data at input byte 0",
		},
		{
			name: "ERROR: invalid EncryptionPublicKeyIV",
			secret: config.CreditcardCoreProcessorSecret{
				EncryptionPublicKeySecret: validKEK,
				EncryptionPublicKeyIV:     "!!!not-valid-base64!!!",
			},
			setupMock: func() {
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(validDataJSON), http.StatusOK, nil)
				logger.On("Error", mock.Anything, "Failed to decrypt card encryption public key", mock.Anything).Once().Return()
			},
			wantErrContains: "decode iv: illegal base64 data at input byte 0",
		},
		{
			name:   "ERROR: data field not valid std base64",
			secret: validSecret,
			setupMock: func() {
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(`{"data":"!!!not-valid-base64!!!"}`), http.StatusOK, nil)
				logger.On("Error", mock.Anything, "Failed to decrypt card encryption public key", mock.Anything).Once().Return()
			},
			wantErrContains: "std encoding for decode public key: illegal base64 data at input byte 0",
		},
		{
			name:   "ERROR: data decodes to invalid URL base64",
			secret: validSecret,
			setupMock: func() {
				// "Ky8rLysv" std-decodes to "+/+/+/" which contains chars invalid in URL base64
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(`{"data":"Ky8rLysv"}`), http.StatusOK, nil)
				logger.On("Error", mock.Anything, "Failed to decrypt card encryption public key", mock.Anything).Once().Return()
			},
			wantErrContains: "url encoding for decode public key: illegal base64 data at input byte 0",
		},
		{
			name:   "ERROR: DecryptAESCBC fails",
			secret: validSecret,
			setupMock: func() {
				httpClient.On("GET", mock.Anything, mock.Anything, mock.Anything).
					Return([]byte(validDataJSON), http.StatusOK, nil)
				crypto.On("DecryptAESCBC", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to decrypt card encryption public key", mock.Anything).Once().Return()
			},
			wantErrContains: assert.AnError.Error(),
		},
		{
			name:   "SUCCESS",
			secret: validSecret,
			setupMock: func() {
				crypto.On("DecryptAESCBC", mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte("public-key-bytes"), nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			secret.CreditcardCoreProcessorSecret = tc.secret

			result, err := repo.GetCardEncryptionPublicKey(t.Context(), "merchant-id")
			if tc.wantErrContains != "" {
				if assert.Error(t, err) {
					assert.ErrorContains(t, err, tc.wantErrContains)
				}

			} else {
				assert.NoError(t, err)
				assert.Equal(t, []byte("public-key-bytes"), result)
			}
			crypto.AssertExpectations(t)
			httpClient.AssertExpectations(t)
		})
	}
}
