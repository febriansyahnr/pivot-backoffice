package snapCoreRepository

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/paper-indonesia/pivot-backoffice/config"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSnapCoreRepository_GenerateQrMpm(t *testing.T) {
	successResponse := `{
		"code": "SNP-CR-00",
		"message": "OK",
		"data": {
			"uuid": "a9f66a8a-8ecf-4bc1-bf83-b45bbe16945c",
			"responseCode": "2004700",
			"responseMessage": "Successful",
			"partnerReferenceNo": "QR1721058341",
			"referenceNo": "2024071522681975162825436",
			"qrContent": "00020101021126740025ID.CO.BANKNEOCOMMERCE.WWW011893600490591008051102120005700009290303UME51550025ID.CO.BANKNEOCOMMERCE.WWW0215BNC2407015535610303UME5204152053033605802ID5906Harsya6012MAMUJU UTARA6105915716242052230018076758519533854720712213141251124630434CA",
			"merchantID": "000550000927",
			"storeID": "000570000929",
			"qrUrl": "https://sit-marketing-img.bankneo.co.id/qris/merchant/img/hkw3-JvkTKEP1lOJPfzgqFzp42_y2RxHlO5DKTl4dak.png",
			"validityPeriod": 0,
			"amount": {
				"value": "",
				"currency": ""
			},
			"createdAt": "2024-07-15T22:45:41.919941+07:00",
			"expiredAt": "2024-07-15T22:45:41.919948+07:00",
			"additionalInfo": {
				"scID": "3001807675851953385472",
				"traceId": "068c3cc2f7b677af3689d77ec5002d0d"
			}
		}
	}`

	testCases := []struct {
		name      string
		wantError bool
		errType   string
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: generate qr mpm to snap core",
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
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			errType:   httpResponse.HttpErrRequest,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)

			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarsal}`), 200, nil)

			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, errors.New("error others"))

			},
		},
		{
			name:      "ERROR: got 400 with field error json from snap core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":{"name":"invalid name"}}`), 400, nil)

			},
		},
		{
			name:      "ERROR: got 500 with field error json from snap core",
			wantError: true,
			errType:   httpResponse.HttpErrThirdParty,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":{"name":"invalid name"}}`), 500, nil)

			},
		},
		{
			name:      "ERROR: got 429 too many requests from snap core",
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
			name:      "ERROR: got 408 request timeout from snap core",
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
			name:      "ERROR: got 502 bad gateway from snap core",
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
			name:      "ERROR: got 503 service unavailable from snap core",
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
			name:      "ERROR: got 504 gateway timeout from snap core",
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

			conf := &config.Config{SnapCoreConfig: config.SnapCoreConfig{BaseUrl: ""}}
			secret := &config.Secret{SnapCoreSecret: struct {
				InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
			}{InternalServiceKey: ""}}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.GenerateQrMpm(context.Background(), snapCoreModel.GenerateQrMpmRequest{})
			if tc.wantError {
				assert.Error(t, err)
				if tc.errType != "" {
					extractedErrType, _ := pkgErr.ExtractError(err)
					assert.Equal(t, tc.errType, extractedErrType)
				}
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)

		})
	}
}
