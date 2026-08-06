package credential_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/credential"
	rabbitMqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	log, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	rmq := rabbitMqMock.NewRabbitMQExt(t)
	rmq.On(
		"PublishActivity", constant.ValueCtxMockType(),
		constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("map[string]string"),
	).Return(nil)

	repo := repositoryMocks.NewICredentialRepository(t)

	request := &credential.CredentialDashboardReq{
		Info: &http.Request{},
	}
	service := New(log, repo, rmq)

	tests := []struct {
		name         string
		mockModifier func()
		wantErr      string
		wantResult   *credential.CredentialDashboardResp
	}{
		{
			name: "ERROR:Some internal error",
			mockModifier: func() {
				repo.On(
					"Get", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:Data not found",
			mockModifier: func() {
				repo.On(
					"Get", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: constant.ErrDataNotFound.Error(),
		},
		{
			name: "SUCCESS",
			mockModifier: func() {
				repo.On(
					"Get", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&credential.CredentialDashboard{
					ClientID:      "unique-client-id",
					ClientSecrets: []credential.ClientSecretSummary{{}},
				}, nil)

			},
			wantResult: &credential.CredentialDashboardResp{
				ClientID: "unique-client-id",
				ClientSecrets: []credential.ClientSecretSummary{{
					KeyName: "Client Secret 1",
				}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockModifier()

			if res, err := service.Get(context.Background(), request); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, res)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
