package snapCoreRepository

import (
	"context"
	"errors"
	"net/http"
	"testing"

	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/virtualAccount"
	httpExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBlockVirtualAccount(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	httpExt := httpExtMock.NewIHTTPRequest(t)

	repo := New(&config.Config{}, &config.Secret{}, logger, httpExt)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []*snapCoreModel.BlockVirtualAccountResponseData
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return(nil, 0, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Unmarshal response", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`A`), 500, nil)
			},
			wantErr: c.ErrInvalidUnmarshalJSON,
		},
		{
			name: "ERROR:Response code 5xx", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"99", "error":{"message":"response code 5xx"}}`), http.StatusInternalServerError, nil)
			},
			wantErr: pkgErrors.New(httpResponse.HttpErrInternal, errors.New("response code 5xx")),
		},
		{
			name: "ERROR:Response code 4xx", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"44", "error":{"message":"response code 4xx"}}`), http.StatusBadRequest, nil)
			},
			wantErr: pkgErrors.New(httpResponse.HttpErrInternal, errors.New("response code 4xx")),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"00", "data":[{"uuid":"123e4567-e89b-12d3-a456-426614174000","acquirer":"BCA","number":"1234567890123456","accountName":"Test Account"}]}`), http.StatusOK, nil)
			},
			wantResult: []*snapCoreModel.BlockVirtualAccountResponseData{
				{
					UUID:        "123e4567-e89b-12d3-a456-426614174000",
					Acquirer:    "BCA",
					Number:      "1234567890123456",
					AccountName: "Test Account",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.BlockVirtualAccount(context.Background(), &snapCoreModel.BlockVirtualAccountRequest{
				MerchantID: "test-merchant-id",
			})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestUnblockVirtualAccount(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	httpExt := httpExtMock.NewIHTTPRequest(t)

	repo := New(&config.Config{}, &config.Secret{}, logger, httpExt)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []*snapCoreModel.UnblockVirtualAccountResponseData
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return(nil, 0, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Unmarshal response", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`A`), 500, nil)
			},
			wantErr: c.ErrInvalidUnmarshalJSON,
		},
		{
			name: "ERROR:Response code 5xx", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"99", "error":{"message":"response code 5xx"}}`), http.StatusInternalServerError, nil)
			},
			wantErr: pkgErrors.New(httpResponse.HttpErrInternal, errors.New("response code 5xx")),
		},
		{
			name: "ERROR:Response code 4xx", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"44", "error":{"message":"response code 4xx"}}`), http.StatusBadRequest, nil)
			},
			wantErr: pkgErrors.New(httpResponse.HttpErrInternal, errors.New("response code 4xx")),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"00", "data":[{"uuid":"123e4567-e89b-12d3-a456-426614174000","acquirer":"BCA","number":"1234567890123456","accountName":"Test Account", "hasError":false}]}`), http.StatusOK, nil)
			},
			wantResult: []*snapCoreModel.UnblockVirtualAccountResponseData{
				{
					Acquirer:    "BCA",
					Number:      "1234567890123456",
					AccountName: "Test Account",
					HasError:    false,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.UnblockVirtualAccount(context.Background(), &snapCoreModel.UnblockVirtualAccountRequest{
				MerchantID: "test-merchant-id",
			})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
