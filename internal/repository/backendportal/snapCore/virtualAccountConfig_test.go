package snapCoreRepository

import (
	"context"
	"errors"
	"testing"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/virtualAccount"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateVirtualAccountConfig(t *testing.T) {
	successResponse := `{
		  "code": "SNP-CR-01",
		  "message": "Created",
		  "data": {
			"uuid": "01969ed1-a31d-76e9-8349-70671b051c48",
			"merchant_id": "4d22c7ed-23ea-42a3-ad81-e7c3a3dd58ea",
			"mid": "1014",
			"bin_prefix": "1480",
			"bin_min": 8,
			"bin_max": 12,
			"type": "CLOSE_STATIC",
			"integration_type": "FACILITATOR",
			"integration_method": "SERVER",
			"client_id": "",
			"credential": {},
			"acquirer": "BCA",
			"status": "ACTIVE"
		  }
		}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: create VA config to snap core",
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
			name:      "ERROR: got 500 with field error json from snap core",
			wantError: true,
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
			_, err := repo.CreateVirtualAccountConfig(context.Background(), &snapCoreModel.CreateVirtualAccountConfigRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)

		})
	}
}

func TestGetVirtualAccountConfig(t *testing.T) {
	successResponse := `{
		  "code": "SNP-CR-01",
		  "message": "Created",
		  "data": [{
			"uuid": "01969ed1-a31d-76e9-8349-70671b051c48",
			"merchant_id": "4d22c7ed-23ea-42a3-ad81-e7c3a3dd58ea",
			"mid": "1014",
			"bin_prefix": "1480",
			"bin_min": 8,
			"bin_max": 12,
			"type": "CLOSE_STATIC",
			"integration_type": "FACILITATOR",
			"integration_method": "SERVER",
			"client_id": "",
			"credential": {},
			"acquirer": "BCA",
			"status": "ACTIVE"
		  }]
		}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: get VA config to snap core",
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
		{
			name:      "SUCCESS: get VA config with valid metadata",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				// Metadata must be present and non-null to trigger Valid=true
				responseWithMetadata := `{
					"code": "SNP-CR-01",
					"message": "OK",
					"data": [{
						"uuid": "01969ed1-a31d-76e9-8349-70671b051c48",
						"merchant_id": "4d22c7ed-23ea-42a3-ad81-e7c3a3dd58ea",
						"mid": "1014",
						"bin_prefix": "1480",
						"bin_min": 8,
						"bin_max": 12,
						"type": "CLOSE_STATIC",
						"integration_type": "FACILITATOR",
						"integration_method": "SERVER",
						"client_id": "",
						"credential": {},
						"acquirer": "BCA",
						"status": "ACTIVE",
						"metadata": {
							"merchant_prefix": {
								"start_range": "001",
								"end_range": "999"
							}
						}
					}]
				}`
				mockHttp.On(
					"GET",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(responseWithMetadata), 200, nil)
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

			// For the metadata test case, set up the test hook
			if tc.name == "SUCCESS: get VA config with valid metadata" {
				repo.testVAConfigPostProcessor = func(data interface{}) {
					if configs, ok := data.([]*snapCoreModel.VirtualAccountConfigResponseData); ok {
						for _, config := range configs {
							// Manually set Metadata.Valid = true and populate JSONText
							config.Metadata.Valid = true
							config.Metadata.JSONText = []byte(`{"merchant_prefix":{"start_range":"001","end_range":"999"}}`)
						}
					}
				}
			}

			_, err := repo.GetVirtualAccountConfig(context.Background(), &snapCoreModel.GetVirtualAccountConfigRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)

		})
	}
}

func TestUpdateVirtualAccountConfigPrefix(t *testing.T) {
	successResponse := `{
		  "code": "SNP-CR-01",
		  "message": "Created",
		  "data": {
			"uuid": "01969ed1-a31d-76e9-8349-70671b051c48",
			"merchant_id": "4d22c7ed-23ea-42a3-ad81-e7c3a3dd58ea",
			"mid": "1014",
			"bin_prefix": "1480",
			"bin_min": 8,
			"bin_max": 12,
			"type": "CLOSE_STATIC",
			"integration_type": "FACILITATOR",
			"integration_method": "SERVER",
			"client_id": "",
			"credential": {},
			"acquirer": "BCA",
			"status": "ACTIVE"
		  }
		}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: update VA config to snap core",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PATCH",
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
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PATCH",
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
					"PATCH",
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
					"PATCH",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, errors.New("error others"))

			},
		},
		{
			name:      "ERROR: got 500 with field error json from snap core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PATCH",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
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
			err := repo.UpdateVirtualAccountConfigPrefix(context.Background(), &snapCoreModel.UpdateVirtualAccountConfigPrefixRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)

		})
	}
}
