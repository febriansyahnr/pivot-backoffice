package snapCoreRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/paper-indonesia/pivot-backoffice/config"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSnapCoreRepository_QueryQrMpmDynamic(t *testing.T) {
	successResponse := `{
		"code": "SNP-CR-00",
		"message": "OK",
		"data": {
			"uuid": "d95cba3f-e67b-4fda-bd2a-4414c7e93310",
			"partnerReferenceNo": "QR1721181254",
			"acquirerReferenceNo": "2024071708161886791625625",
			"qrContent": "00020101021226740025ID.CO.BANKNEOCOMMERCE.WWW011893600490591008051102120005100009280303UBE51550025ID.CO.BANKNEOCOMMERCE.WWW0215BNC2407013497200303UBE5204078053033605405100005802ID5910sub Harsya6006SIDRAP6105916146233012230018133917344703242250703A01630498C0",
			"merchantID": "000550000927",
			"storeID": "000570000929",
			"qrUrl": "https://sit-marketing-img.bankneo.co.id/qris/merchant/img/VkYRk4-FXSEVgLJV5hVQByDD46if2PLIVpR3F5OJFNs.png",
			"validityPeriod": 8000,
			"status": "CANCELED",
			"qrType": "DYNAMIC",
			"amount": {
				"value": "10000",
				"currency": "IDR"
			},
			"feeAmount": {
				"value": "",
				"currency": ""
			},
			"createdAt": "2024-07-17T01:54:15Z",
			"expiredAt": "2024-07-17T04:07:35Z",
			"additionalInfo": null
		}
	}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: query qr mpm dynamic to snap core",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
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
					mock.AnythingOfType("string"),
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
					mock.AnythingOfType("string"),
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
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, errors.New("error others"))

			},
		},
		{
			name:      "ERROR: got 400 with field error json from snap core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":{"name":"invalid name"}}`), 400, nil)

			},
		},
		{
			name:      "ERROR: got 500 with field error json from snap core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":{"name":"invalid name"}}`), 500, nil)

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
			_, err := repo.QueryQrMpmDynamic(context.Background(), "uuid")
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)

		})
	}
}
