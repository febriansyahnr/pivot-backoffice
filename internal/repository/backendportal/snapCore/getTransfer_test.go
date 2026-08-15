package snapCoreRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransferById(t *testing.T) {
	successResponse := `{
		"data": {
			"responseCode": "200",
			"responseMessage": "success",
			"uuid": "04918815-e291-4930-a44e-509dcf1873ed",
			"partnerReferenceNo": "1234",
			"bankReferenceNo": "5678",
			"amount": {
				"value": "10000.50",
				"currency": "IDR"
			},
			"beneficiaryAccountNo": "123412341234",
			"beneficiaryBankCode": "008",
			"sourceAccountNo": "987698769876",
			"status": "SUCCESS",
			"transferType": "INTRABANK",
			"externalID": "ext-123"
		}
	}`

	testCases := []struct {
		name      string
		externalId string
		wantError bool
		wantNil   bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:       "SUCCESS: Get transfer by ID with valid response",
			externalId: "external-id-123",
			wantError:  false,
			wantNil:    false,
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
			name:       "ERROR: HTTP request error - response is nil",
			externalId: "external-id-123",
			wantError:  true,
			wantNil:    true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, errors.New("network error"))
			},
		},
		{
			name:       "SUCCESS: Response with error but data is returned",
			externalId: "external-id-123",
			wantError:  true,
			wantNil:    false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","data":{"uuid":"test-uuid","status":"FAILED"}}`), 400, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{SnapCoreConfig: config.SnapCoreConfig{BaseUrl: ""}}
			secret := &config.Secret{SnapCoreSecret: struct {
				InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
			}{InternalServiceKey: ""}}

			repo := New(conf, secret, mockLogger, mockHttp)
			result, err := repo.GetTransferById(context.Background(), tc.externalId, false)

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}

			mockHttp.AssertExpectations(t)
		})
	}
}
