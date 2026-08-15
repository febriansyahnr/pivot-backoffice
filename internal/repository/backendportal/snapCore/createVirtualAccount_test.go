package snapCoreRepository

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/virtualAccount"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSnapCoreRepository_CreateVirtualAccount(t *testing.T) {
	successResponse := `{
		"customerNo": "04918815-e291-4930-a44e-509dcf1873ed",
		"accountName": "Febri Close Static",
		"accountEmail": "febri@mail.co.id",
		"accountPhone": "08123456789",
		"subCompany": "paper-sub",
		"totalAmount": {
			"value": "10000.50",
			"currency": "IDR"
		},
		"feeAmount": {
			"value": "10000.50",
			"currency": "IDR"
		},
		"billDetails": [
			{
				"billerReferenceId": "ref123",
				"billCode": "ELEC",
				"billNo": "1234",
				"billName": "Electricity Bill",
				"billShortName": "Electricity",
				"billDescription": {
					"en": "Monthly electricity bill",
					"id": "Tagihan listrik bulanan"
				},
				"billSubCompany": "PLN",
				"billAmount": {
					"amount": "100000.00",
					"currency": "IDR"
				},
				"additionalInfo": {
					"period": "Jan 2022"
				}
			}
		],
		"freeTexts": [
			{
				"english": "Hello World",
				"indonesia": "Hallo Dunia"
			}
		],
		"acquirer": "permata",
		"vaNumber": "00000003",
		"mid": "1234",
		"isClosedAmount": true,
		"isSingleUse": false
	}`

	testCases := []struct {
		name      string
		wantError bool
		errType   string
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: create VA to snap core",
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
			_, err := repo.CreateVirtualAccount(context.Background(), snapCoreModel.CreateVirtualAccountRequest{})
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
