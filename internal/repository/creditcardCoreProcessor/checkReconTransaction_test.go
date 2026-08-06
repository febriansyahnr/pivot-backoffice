package creditcardCoreProcessorRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreditcardCoreProcessorRepositoryCheckReconTransaction(t *testing.T) {
	successResponse := `{
    "code": "CC-CR-00",
    "message": "OK",
    "data": {
        "uuid": "d7f09d65-0377-4630-be89-317ba62d681f",
        "type": "CARD",
        "status": "SUCCESS",
        "code": "OK"
    }
}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: check recon transaction",
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
			name:      "ERROR: HTTP request failed",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return(nil, 0, errors.New("http request failed"))
			},
		},
		{
			name:      "ERROR: Status code 400",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				errorResponse := `{
                    "code": "CC-CR-400",
                    "message": "Bad Request"
                }`
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(errorResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: Status code 500",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				errorResponse := `{
                    "code": "CC-CR-500",
                    "message": "Internal Server Error"
                }`
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(errorResponse), 500, nil)
			},
		},
		{
			name:      "ERROR: Invalid JSON response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				invalidResponse := `{invalid json`
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(invalidResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: Response with error object (400)",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				errorResponse := `{
                    "code": "CC-CR-400",
                    "message": "Validation Error",
                    "error": {
                        "field": "type",
                        "message": "type is required"
                    }
                }`
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(errorResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: Response with error object (500)",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				errorResponse := `{
                    "code": "CC-CR-500",
                    "message": "Database Error",
                    "error": {
                        "detail": "connection timeout",
                        "code": "DB_TIMEOUT"
                    }
                }`
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(errorResponse), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHttp := httpMocks.NewIHTTPRequest(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			repo := &creditcardCoreProcessorRepository{
				config: &config.Config{
					CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
						BaseUrl: "https://test-cc-processor.com",
					},
				},
				secret: &config.Secret{
					CreditcardCoreProcessorSecret: config.CreditcardCoreProcessorSecret{
						InternalServiceKey: "test-key",
					},
				},
				logger:      mockLogger,
				httpRequest: mockHttp,
			}

			tc.setupMock(mockHttp)

			request := &creditcardCoreProcessorModel.AutoReconTrxRequest{
				Type:        "CARD",
				ReferenceNo: "test-ref",
			}

			result, err := repo.CheckReconTransaction(context.Background(), request)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "d7f09d65-0377-4630-be89-317ba62d681f", result.UUID)
				assert.Equal(t, "CARD", result.Type)
				assert.Equal(t, "SUCCESS", result.Status)
				assert.Equal(t, constant.ReconCCCodeOk, result.Code)
			}
		})
	}
}
