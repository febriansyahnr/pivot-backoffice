package xbCoreProcessorRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutById(t *testing.T) {
	successResponse := `{
		"code": "00",
		"data": {
			"reference_number": "REF001",
			"payout_id": "48e0d7dd-c10f-4032-a70f-64357ee34939",
			"status": "EXPIRED",
			"autobookfx_confirmation_id": "",
			"sub_status": "",
			"status_description": "Salary",
			"created_at": "2024-09-03T17:23:05Z",
			"updated_at": "2024-09-04T00:43:38Z"
		}
	}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "ERROR: error when do request get payout by id",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("GET", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType(constant.MockTypeMapStringStringReference)).Return(nil, 0, errors.New("error when do request get payout by id"))
			},
		},
		{
			name:      "ERROR: error when read get payout by id response body",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("GET", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType(constant.MockTypeMapStringStringReference)).Return([]byte("{[}"), 0, nil)
			},
		},
		{
			name:      "ERROR: got error 400 when get payout by id",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("GET", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType(constant.MockTypeMapStringStringReference)).Return([]byte(`{"error": "Unprocessable Entity"}`), 422, nil)
			},
		},
		{
			name:      "ERROR: got error 500 when get payout by id",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("GET", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType(constant.MockTypeMapStringStringReference)).Return([]byte(`{"error": "Internal Server Error"}`), 500, nil)
			},
		},
		{
			name:      "SUCCESS: get payout by id",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("GET", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType(constant.MockTypeMapStringStringReference)).Return([]byte(successResponse), 200, nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tt.setupMock(mockHttp)

			conf := &config.Config{
				XbCoreProcessorConfig: config.XbCoreProcessorConfig{
					BaseUrl: "http://localhost:8080",
				},
			}
			secret := &config.Secret{
				XbCoreProcessorSecret: config.XbCoreProcessorSecret{
					InternalServiceKey: "INTERNAL_SERVICE_KEY",
				},
			}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.GetPayoutById(context.TODO(), &xbCoreProcessorModel.GetPayoutRequest{
				Id:         "48e0d7dd-c10f-4032-a70f-64357ee34939",
				MerchantId: "MERCHANT_ID",
			})
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)

		})
	}
}
