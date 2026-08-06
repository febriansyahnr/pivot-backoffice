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

func TestGetRfiDetails(t *testing.T) {
	successResponse := `{
		"code": "00",
		"data": [
			{
				"uuid": "d501227c-c764-438e-ae92-b1cf74031f8e",
				"payout_id": "6a003098-b587-4c13-8dfd-9d91d21e892c",
				"partner_document_id": "66d818e1e100a8f852a50dd7",
				"partner_document_entity_id": "66d824dec667b38349fd6d45",
				"actor": "BENEFICIARY",
				"entity": "PASSPORT",
				"document_type": "FILE",
				"filename": "",
				"location": {},
				"value": "FILE",
				"comment": "test",
				"status": "received",
				"requested_at": "2024-09-04T09:14:07Z",
				"created_at": "2024-09-05T06:29:53Z",
				"updated_at": "2024-09-07T11:10:42Z"
			}
		]
	}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "ERROR: error when do request get rfi details",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("GET", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType(constant.MockTypeMapStringStringReference)).Return(nil, 0, errors.New("error when do request get rfi details"))
			},
		},
		{
			name:      "ERROR: error when read get rfi details response body",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("GET", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType(constant.MockTypeMapStringStringReference)).Return([]byte("{[}"), 0, nil)
			},
		},
		{
			name:      "ERROR: got error 400 when get rfi details",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("GET", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType(constant.MockTypeMapStringStringReference)).Return([]byte(`{"error": "Unprocessable Entity"}`), 422, nil)
			},
		},
		{
			name:      "ERROR: got error 500 when get rfi details",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("GET", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType(constant.MockTypeMapStringStringReference)).Return([]byte(`{"error": "Internal Server Error"}`), 500, nil)
			},
		},
		{
			name:      "SUCCESS: get rfi details",
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
			_, err := repo.GetRfiDetails(context.TODO(), &xbCoreProcessorModel.GetRfiDetailsRequest{
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
