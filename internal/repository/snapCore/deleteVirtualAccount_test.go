package snapCoreRepository

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	httpExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteVirtualAccount(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	httpExt := httpExtMock.NewIHTTPRequest(t)

	repo := New(&config.Config{}, &config.Secret{}, logger, httpExt)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *snapCoreModel.DeleteVirtualAccountResponseData
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"DELETE", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return(nil, 0, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Unmarshal response", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"DELETE", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`A`), 500, nil)
			},
			wantErr: c.ErrInvalidUnmarshalJSON,
		},
		{
			name: "ERROR:Response code 5xx", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"DELETE", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"99", "error":{"message":"response code 5xx"}}`), http.StatusInternalServerError, nil)
			},
			wantErr: pkgErrors.New(httpResponse.HttpErrInternal, errors.New("response code 5xx")),
		},
		{
			name: "ERROR:Response code 4xx", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"DELETE", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"44", "error":{"message":"response code 4xx"}}`), http.StatusBadRequest, nil)
			},
			wantErr: pkgErrors.New(httpResponse.HttpErrInternal, errors.New("response code 4xx")),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"DELETE", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"00", "data":{"uuid":"75a56d3f-8bb8-476b-9662-4a7e9f843484"}}`), http.StatusOK, nil)
			},
			wantResult: &snapCoreModel.DeleteVirtualAccountResponseData{UUID: "75a56d3f-8bb8-476b-9662-4a7e9f843484"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.DeleteVirtualAccount(context.Background(), &snapCoreModel.DeleteVirtualAccountRequest{})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
