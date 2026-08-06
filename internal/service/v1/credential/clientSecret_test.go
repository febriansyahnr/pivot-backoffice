package credential_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/credential"
	rabbitMqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClientSecretById(t *testing.T) {
	traceID := uuid.NewString()

	buf := new(bytes.Buffer)
	defer buf.Reset()

	ctx := context.WithValue(t.Context(), pdkConst.CtxTraceIdKey, traceID)
	log := loggerMock.NewSlogger(loggerMock.Config{}, loggerMock.WithSlogOutput(buf))

	vaultTransit := vaultMock.NewIVaultTransit(t)

	rmq := rabbitMqMock.NewRabbitMQExt(t)
	rmq.On(
		"PublishActivity", constant.ValueCtxMockType(),
		constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
	).Return(nil)
	ptrClientSecretMockType := mock.AnythingOfType("*credential.ClientSecret")

	userService := serviceMocks.NewIUserService(t)
	repo := repositoryMocks.NewICredentialRepository(t)

	request := &credential.ClientSecretReq{
		Info: &http.Request{},
	}
	requestWithActionGet := &credential.ClientSecretReq{
		Info:   &http.Request{},
		Action: http.MethodGet,
	}
	requestWithActionPost := &credential.ClientSecretReq{
		Info:   &http.Request{},
		Action: http.MethodPost,
	}

	service := New(log, repo, rmq, WithUserService(userService), WithVaultTransit(vaultTransit))

	tests := []struct {
		name         string
		request      *credential.ClientSecretReq
		mockModifier func()
		wantErr      string
	}{
		{
			name: "ERROR:Validate PIN",
			mockModifier: func() {
				userService.On(
					"CheckCurrentPin", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:Invalid PIN",
			mockModifier: func() {
				userService.On(
					"CheckCurrentPin", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidPIN))
			},
			wantErr: "invalid pin",
		},
		{
			name:    "ERROR:Get Secret/Get client secret by id",
			request: requestWithActionGet,
			mockModifier: func() {
				userService.On(
					"CheckCurrentPin", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)

				repo.On(
					"GetClientSecretById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(constant.InternalErrorFmt, traceID),
		},
		{
			name:    "ERROR:Get Secret/Client secret not found",
			request: requestWithActionGet,
			mockModifier: func() {
				repo.On(
					"GetClientSecretById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "data not found",
		},
		{
			name:    "ERROR:Encrypt merchant secret",
			request: requestWithActionPost,
			mockModifier: func() {
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser).Error(),
		},
		{
			name:    "ERROR:Generate Secret/Update client secret",
			request: requestWithActionPost,
			mockModifier: func() {
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Return(&vault.EncryptResponse{}, nil)
				repo.On(
					"UpdateClientSecretById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), ptrClientSecretMockType,
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(constant.InternalErrorFmt, traceID),
		},
		{
			name:    "ERROR:Generate Secret/Data not found",
			request: requestWithActionPost,
			mockModifier: func() {
				repo.On(
					"UpdateClientSecretById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), ptrClientSecretMockType,
				).Once().Return(false, nil)
			},
			wantErr: "data not found",
		},
		{
			name:    "ERROR:Decrypt merchant secret",
			request: requestWithActionGet,
			mockModifier: func() {
				repo.On(
					"GetClientSecretById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(&credential.ClientSecret{Secret: "ABC", SecretVersion: 1}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser).Error(),
		},
		{
			name:    "SUCCESS:Get Secret",
			request: requestWithActionGet,
			mockModifier: func() {
				repo.On(
					"GetClientSecretById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(&credential.ClientSecret{Secret: "ABC", SecretVersion: 1}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Return(&vault.DecryptResponse{Plaintext: []byte(`ABC`)}, nil)
			},
		},
		{
			name:    "SUCCESS:Generate Secret",
			request: requestWithActionPost,
			mockModifier: func() {
				repo.On(
					"UpdateClientSecretById", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), ptrClientSecretMockType,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.request == nil {
				test.request = request
			}
			test.mockModifier()

			if res, err := service.ClientSecretById(ctx, test.request); test.wantErr == "" {
				require.NoError(t, err)
				assert.NotEmpty(t, res.Secret)
				assert.Greater(t, res.Time, time.Now().Add(-time.Minute).UTC().Unix())

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
