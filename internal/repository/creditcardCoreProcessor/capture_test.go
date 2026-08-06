package creditcardCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	httpRequestMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCapture(t *testing.T) {
	tests := []struct {
		name      string
		request   *creditcardCoreProcessorModel.CaptureRequest
		setupMock func(*httpRequestMocks.IHTTPRequest)
		wantErr   bool
		wantData  *creditcardCoreProcessorModel.CaptureResponseData
	}{
		{
			name: "SUCCESS",
			request: &creditcardCoreProcessorModel.CaptureRequest{
				MerchantID:            "merchant-123",
				ClientTransactionID:   "txn-123",
				AcquirerTransactionID: "acq-123",
				Currency:              "IDR",
				Amount:                100000,
			},
			setupMock: func(mockHTTP *httpRequestMocks.IHTTPRequest) {
				successResp := creditcardCoreProcessorModel.CaptureResponse{
					Code:    "00",
					Message: "Success",
					Data: creditcardCoreProcessorModel.CaptureResponseData{
						ID:                    "txn-123",
						Status:                "CAPTURED",
						AcquirerTransactionID: "acq-123",
						Currency:              "IDR",
						Amount:                100000,
					},
				}
				respBytes, _ := json.Marshal(successResp)
				mockHTTP.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(respBytes, 200, nil).Once()
			},
			wantErr: false,
			wantData: &creditcardCoreProcessorModel.CaptureResponseData{
				ID:     "txn-123",
				Status: "CAPTURED",
			},
		},
		{
			name: "ERROR: HTTP request failed",
			request: &creditcardCoreProcessorModel.CaptureRequest{
				MerchantID:          "merchant-123",
				ClientTransactionID: "txn-123",
				Amount:              100000,
			},
			setupMock: func(mockHTTP *httpRequestMocks.IHTTPRequest) {
				mockHTTP.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, 0, errors.New("connection error")).Once()
			},
			wantErr:  true,
			wantData: nil,
		},
		{
			name: "ERROR: Invalid JSON response",
			request: &creditcardCoreProcessorModel.CaptureRequest{
				MerchantID:          "merchant-123",
				ClientTransactionID: "txn-123",
				Amount:              100000,
			},
			setupMock: func(mockHTTP *httpRequestMocks.IHTTPRequest) {
				mockHTTP.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]byte("invalid json"), 200, nil).Once()
			},
			wantErr:  true,
			wantData: nil,
		},
		{
			name: "ERROR: 400 status code",
			request: &creditcardCoreProcessorModel.CaptureRequest{
				MerchantID:          "merchant-123",
				ClientTransactionID: "txn-123",
				Amount:              100000,
			},
			setupMock: func(mockHTTP *httpRequestMocks.IHTTPRequest) {
				errorResp := creditcardCoreProcessorModel.CaptureResponse{
					Code:    "400",
					Message: "Bad Request",
				}
				respBytes, _ := json.Marshal(errorResp)
				mockHTTP.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(respBytes, 400, nil).Once()
			},
			wantErr:  true,
			wantData: nil,
		},
		{
			name: "ERROR: 400 status code with error details",
			request: &creditcardCoreProcessorModel.CaptureRequest{
				MerchantID:          "merchant-123",
				ClientTransactionID: "txn-123",
				Amount:              100000,
			},
			setupMock: func(mockHTTP *httpRequestMocks.IHTTPRequest) {
				errorResp := creditcardCoreProcessorModel.CaptureResponse{
					Code:    "400",
					Message: "Bad Request",
					Error: map[string]string{
						"type":    "VALIDATION_ERROR",
						"message": "Invalid capture amount",
					},
				}
				respBytes, _ := json.Marshal(errorResp)
				mockHTTP.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(respBytes, 400, nil).Once()
			},
			wantErr:  true,
			wantData: nil,
		},
		{
			name: "ERROR: 500 status code",
			request: &creditcardCoreProcessorModel.CaptureRequest{
				MerchantID:          "merchant-123",
				ClientTransactionID: "txn-123",
				Amount:              100000,
			},
			setupMock: func(mockHTTP *httpRequestMocks.IHTTPRequest) {
				errorResp := creditcardCoreProcessorModel.CaptureResponse{
					Code:    "500",
					Message: "Internal Server Error",
				}
				respBytes, _ := json.Marshal(errorResp)
				mockHTTP.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(respBytes, 500, nil).Once()
			},
			wantErr:  true,
			wantData: nil,
		},
		{
			name: "ERROR: 500 status code with error details",
			request: &creditcardCoreProcessorModel.CaptureRequest{
				MerchantID:          "merchant-123",
				ClientTransactionID: "txn-123",
				Amount:              100000,
			},
			setupMock: func(mockHTTP *httpRequestMocks.IHTTPRequest) {
				errorResp := creditcardCoreProcessorModel.CaptureResponse{
					Code:    "500",
					Message: "Internal Server Error",
					Error: map[string]string{
						"type":    "DATABASE_ERROR",
						"message": "Database connection failed",
					},
				}
				respBytes, _ := json.Marshal(errorResp)
				mockHTTP.On("POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(respBytes, 500, nil).Once()
			},
			wantErr:  true,
			wantData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTP := httpRequestMocks.NewIHTTPRequest(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.setupMock(mockHTTP)

			cfg := &config.Config{
				CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
					BaseUrl: "https://test.example.com",
				},
			}
			secret := &config.Secret{
				CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
					InternalServiceKey: "test-key",
				},
			}

			repo := New(cfg, secret, mockLogger, mockHTTP)
			ctx := context.Background()

			result, err := repo.Capture(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.wantData != nil {
					assert.Equal(t, tt.wantData.ID, result.ID)
					assert.Equal(t, tt.wantData.Status, result.Status)
				}
			}

			mockHTTP.AssertExpectations(t)
		})
	}
}
