package xbCoreProcessorRepository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/xbCoreProcessor"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestCreateSender(t *testing.T) {
	successResponse := `{"example": "ok"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: create sender to xb core",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
					mock.Anything,
				).Return([]byte(successResponse), 200, nil)

			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
					mock.Anything,
				).Return([]byte(successResponse), 400, nil)

			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
					mock.Anything,
				).Return([]byte(`{error unmarsal}`), 200, nil)

			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
					mock.Anything,
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)

			},
		},
		{
			name:      "ERROR: got 400 with field error json from xb core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
					mock.Anything,
				).Return([]byte(`{"code":"400","error":{"name":"invalid name"}}`), 400, nil)

			},
		},
		{
			name:      "ERROR: got 500 with field error json from xb core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.CreateSenderRequest"),
					mock.Anything,
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
			_, err := repo.CreateSender(context.Background(), &xbCoreProcessorModel.CreateSenderRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)

		})
	}
}
