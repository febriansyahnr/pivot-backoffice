package creditcardCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateEncryptedCardAuthenticationLink(t *testing.T) {
	tests := []struct {
		name           string
		request        *creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest
		mockResponse   []byte
		mockStatusCode int
		mockError      error
		expectedResult *creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse
		expectedError  error
	}{
		{
			name: "success",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
				CardID:              "card-abc",
				CVC:                 "123",
				Amount:              1000.50,
				Fee:                 50.25,
				Currency:            "IDR",
			},
			mockResponse: []byte(`{
				"code": "200",
				"message": "Success",
				"data": {
					"acquirer_transaction_id": "acquirer-123",
					"amount": "1000.50",
					"currency": "IDR",
					"message": "Success",
					"session_id": "session-456",
					"status": "AUTHENTICATED",
					"authentication_url": {
						"action_url": "https://action.url",
						"created_at": "2023-01-01T00:00:00Z",
						"creq": "token-123",
						"html": "<html>3DS form</html>",
						"method": "POST",
						"url": "https://3ds.url",
						"version": "2.1.0"
					}
				}
			}`),
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectedResult: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				AcquirerTransactionID: "acquirer-123",
				Amount:                "1000.50",
				Currency:              "IDR",
				Message:               "Success",
				SessionID:             "session-456",
				Status:                "AUTHENTICATED",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					ActionURL:    "https://action.url",
					CreatedAt:    "2023-01-01T00:00:00Z",
					ThreeDSToken: "token-123",
					HTML:         "<html>3DS form</html>",
					Method:       "POST",
					URL:          "https://3ds.url",
					Version:      "2.1.0",
				},
			},
			expectedError: nil,
		},
		{
			name: "http request error",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-456",
			},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("network error"),
			expectedResult: nil,
			expectedError:  errors.New("network error"),
		},
		{
			name: "json unmarshal error",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-456",
			},
			mockResponse:   []byte(`invalid json`),
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  &json.SyntaxError{},
		},
		{
			name: "server error - 500",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-456",
			},
			mockResponse: []byte(`{
				"code": "500",
				"message": "Internal Server Error"
			}`),
			mockStatusCode: http.StatusInternalServerError,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrInternal, errors.New("Internal Server Error")),
		},
		{
			name: "client error - decryption key does not match",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "decryption key used does not match"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, errors.New("failed to decrypt card payment")),
		},
		{
			name: "client error - failed while decrypting card encryption key",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "failed while decrypting card encryption key"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, errors.New("failed to decrypt card payment")),
		},
		{
			name: "client error - failed while decrypting card details",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "failed while decrypting card details"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, errors.New("failed to decrypt card payment")),
		},
		{
			name: "client error - failed to parse card details in JSON format",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "failed to parse card details in JSON format"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, errors.New("failed to decrypt card payment")),
		},
		{
			name: "client error - invalid card number format",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "invalid card number format"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, errors.New("invalid card number format")),
		},
		{
			name: "client error - invalid card number",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "invalid card number"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, errors.New("invalid card number format")),
		},
		{
			name: "client error - invalid network token cryptogram",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "invalid network token cryptogram"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, errors.New("invalid network token cryptogram")),
		},
		{
			name: "client error - there is ongoing network token transaction for same payment id",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "422",
				"message": "there is ongoing network token transaction for same payment id"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrUnprocessableContent, errors.New("there is ongoing network token transaction for same payment id")),
		},
		{
			name: "client error - invalid card information (default case)",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "invalid expiry date"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, errors.New("invalid card information")),
		},
		{
			name: "client error - billing information required",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "billing information is required for foreign card"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrForeignCardBillingInformationMissing),
		},
		{
			name: "client error - billing information first name required",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "foreign cards require billing information: First Name"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.givenName is required")),
		},
		{
			name: "client error - billing information last name required",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "foreign cards require billing information: Last Name"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.surname is required")),
		},
		{
			name: "client error - billing information address line required",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "foreign cards require billing information: Address Line"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.addressLine is required")),
		},
		{
			name: "client error - billing information city required",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "foreign cards require billing information: City"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.city is required")),
		},
		{
			name: "client error - billing information State/province required",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "foreign cards require billing information: State/Province"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.provinceState is required")),
		},
		{
			name: "client error - billing information Postal Code required",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "foreign cards require billing information: Postal Code"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.postalCode is required")),
		},
		{
			name: "client error - billing information Country required",
			request: &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
			},
			mockResponse: []byte(`{
				"code": "400",
				"message": "foreign cards require billing information: Country"
			}`),
			mockStatusCode: http.StatusBadRequest,
			mockError:      nil,
			expectedResult: nil,
			expectedError:  pkgErrors.New(httpResponse.HttpErrRequest, fmt.Errorf("billingInfo.country is required")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := logger.NewSlogger(logger.Config{})
			httpMocks := &httpMocks.IHTTPRequest{}

			repo := &creditcardCoreProcessorRepository{
				config: &config.Config{
					CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
						BaseUrl: "https://test.url",
					},
				},
				secret: &config.Secret{
					CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
						InternalServiceKey: "key",
					},
				},
				logger:      mockLogger,
				httpRequest: httpMocks,
			}

			ctx := context.Background()

			httpMocks.On("POST", mock.Anything, mock.AnythingOfType("string"), tt.request, mock.AnythingOfType("map[string]string")).
				Return(tt.mockResponse, tt.mockStatusCode, tt.mockError)

			result, err := repo.CreateEncryptedCardAuthenticationLink(ctx, tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.name == "json unmarshal error" {
					assert.IsType(t, &json.SyntaxError{}, err)
				} else if tt.name == "http request error" {
					assert.Equal(t, tt.expectedError.Error(), err.Error())
				} else if tt.name == "server error - 500" {
					assert.Contains(t, err.Error(), "Internal Server Error")
				} else {
					// For all card error cases, verify the error message matches expected
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			httpMocks.AssertExpectations(t)
		})
	}
}

func TestAuthentication(t *testing.T) {
	mockLogger := logMock.NewILogger(t)
	httpClient := httpMocks.NewIHTTPRequest(t)

	repo := New(&config.Config{}, nil, mockLogger, httpClient)

	request := creditcardCoreProcessorModel.AuthenticationRequest{
		MerchantID:       "a8e58da5-ffd4-48b7-aa89-b62aa95335ed", // NOSONAR
		PaymentID:        "a13c4ebb-87c0-4aaa-8f4e-dbc97beaedd0", // NOSONAR
		EncryptedPayload: "encrypted-payload",                    // NOSONAR
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *creditcardCoreProcessorModel.AuthenticationResponse
	}{
		{
			name: "ERROR: HTTP POST fails",
			setupMock: func() {
				httpClient.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(nil, 0, assert.AnError)
				mockLogger.On("Info", mock.Anything, "Failed to send authentication request to card processor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
				mockLogger.On("Error", mock.Anything, "Failed while performing authentication request (encrypted card web view flow)", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR: HTTP status 4xx",
			setupMock: func() {
				httpClient.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(`bad request`), http.StatusBadRequest, nil)
				mockLogger.On("Info", mock.Anything, "Failed to send authentication request to card processor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
				mockLogger.On("Error", mock.Anything, "Failed while performing authentication request (encrypted card web view flow), received http code 400", mock.Anything).Once().Return()
			},
			wantError: pkgErrors.NewNonRetryableError(errors.New("partner response: bad request")),
		},
		{
			name: "ERROR: HTTP status 5xx",
			setupMock: func() {
				httpClient.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(`internal server error`), http.StatusInternalServerError, nil)
				mockLogger.On("Info", mock.Anything, "Failed to send authentication request to card processor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
				mockLogger.On("Error", mock.Anything, "Failed while performing authentication request (encrypted card web view flow), received http code 500", mock.Anything).Once().Return()
			},
			wantError: errors.New("partner response: internal server error"),
		},
		{
			name: "ERROR: unmarshal response fails",
			setupMock: func() {
				httpClient.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(`{invalid json}`), http.StatusOK, nil)
				mockLogger.On("Error", mock.Anything, "Failed while unmarshalling authentication response", mock.Anything).Once().Return()
			},
			wantError: func() error {
				return pkgErrors.NewNonRetryableError(fmt.Errorf("unmarshalling json: %w", json.Unmarshal([]byte(`{invalid json}`), &struct{}{})))
			}(),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				httpClient.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]byte(`{"code":"00","data":{"status":"AUTHENTICATED","session_id":"session-123","acquirer_transaction_id":"acq-456","currency":"IDR"}}`), http.StatusOK, nil)
			},
			wantResult: &creditcardCoreProcessorModel.AuthenticationResponse{
				Status:                "AUTHENTICATED",
				SessionID:             "session-123",
				AcquirerTransactionID: "acq-456",
				Currency:              "IDR",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, err := repo.Authentication(t.Context(), request)
			assert.Equal(t, tc.wantError, err)
			assert.Equal(t, tc.wantResult, result)

			mockLogger.AssertExpectations(t)
			httpClient.AssertExpectations(t)
		})
	}
}
