package snapCoreRepository

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	httpExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelQrMpm(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	httpExt := httpExtMock.NewIHTTPRequest(t)

	repo := New(&config.Config{}, &config.Secret{}, logger, httpExt)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *snapCoreModel.CancelQrMpmResponseData
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
			name: "ERROR:Response code above than 4xx", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"DELETE", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"44", "error":{"message":"response code 4xx"}}`), http.StatusBadRequest, nil)
			},
			wantErr: pkgErrors.New(httpResponse.HttpErrInternal, errors.New("response code 4xx")),
		},
		{
			name: "ERROR:Response code 401", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"DELETE", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return(nil, http.StatusUnauthorized, nil)
			},
			wantErr: pkgErrors.New(httpResponse.HttpErrInternal, constant.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				httpExt.On(
					"DELETE", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"00", "data":{"uuid":"75a56d3f-8bb8-476b-9662-4a7e9f843484"}}`), http.StatusOK, nil)
			},
			wantResult: &snapCoreModel.CancelQrMpmResponseData{UUID: "75a56d3f-8bb8-476b-9662-4a7e9f843484"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.CancelQrMpm(context.Background(), uuid.NewString())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
