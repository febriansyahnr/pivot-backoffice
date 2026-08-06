package user_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/user"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInvitationURL(t *testing.T) {

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	db, clientMock := redismock.NewClientMock()
	userRepo := repoMocks.NewIUserRepository(t)

	traceId := uuid.NewString()
	merchantId := uuid.NewString()
	pattern := "backend-portal:users:user-invitation:token:25e4fcefa9922c8d84d6de5ab0ad665cb47cd33c5bff304268bbc498afdbb8bf*"

	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	cfg := &config.Config{
		MerchantPortalConfig: config.MerchantPortalConfig{
			UserInvitationURL: "http://localhost/invitation",
		},
	}
	service := New(cfg, nil, logger, userRepo, nil, WithRedisClient(redisExt.WrapRedisClient(db, nil)))

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult string
	}{
		{
			name: "ERROR:Find user by email",
			setupMock: func() {
				userRepo.On(
					"FindUserByEmail", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("FU: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:User not found",
			setupMock: func() {
				userRepo.On(
					"FindUserByEmail", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "data not found",
		},
		{
			name: "ERROR:Difference merchant id",
			setupMock: func() {
				userRepo.On(
					"FindUserByEmail", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&user.User{}, nil)
			},
			wantErr: "data not found",
		},
		{
			name: "ERROR:Redis keys",
			setupMock: func() {
				userRepo.On(
					"FindUserByEmail", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&user.User{MerchantId: merchantId}, nil)

				clientMock.ExpectKeys(pattern).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("RD: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Invitation url not found",
			setupMock: func() {
				clientMock.ExpectKeys(pattern).SetVal(nil)
			},
			wantErr: "invitation url not found",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				clientMock.ExpectKeys(pattern).SetVal([]string{"generate-token-invitation"})
			},
			wantResult: "http://localhost/invitation?token=generate-token-invitation",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			clientMock.ClearExpect()

			test.setupMock()

			url, err := service.GetInvitationURL(ctx, merchantId, "email@example.id")
			if test.wantErr == "" {
				require.NoError(t, err)
				assert.NotEmpty(t, url)
				assert.Equal(t, test.wantResult, url)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
