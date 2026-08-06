package user

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisPkgMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func TestSendGeneratedInvitationURL(t *testing.T) {
	defLockKey := "backend-portal:users:email@example.id:user-invitation:lock"
	dataKey := "backend-portal:users:email@example.id:user-invitation:data"
	lastToken := "last-token"
	deleteTokenKey := "backend-portal:users:user-invitation:token:" + lastToken
	userInvitationData := map[string]interface{}{
		constant.UserInvitationTotalResendField: 0,
		constant.UserInvitationLastTokenField:   lastToken,
	}
	feature := constant.UserIdentifierUserInvitation

	rmq := rmqMock.NewRabbitMQExt(t)
	rmq.On(
		"Publish", constant.BackgroundCtxMockType(), constant.StringMockType(), mock.Anything, mock.AnythingOfType("[]uint8"),
	).Return(nil)

	testCases := []struct {
		name       string
		mocksSetup func(

			r redismock.ClientMock,
			limiterMock *redisPkgMocks.ILimiter,
		)
		wantErr  bool
		isResend bool
	}{
		{
			name: "SUCCESS: User invitation with URL",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, constant.StringMockType(), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 1}, nil)

				r.ExpectHGet(dataKey, constant.UserInvitationLastTokenField).SetVal(lastToken)
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)

				r.CustomMatch(func(expected, actual []interface{}) error {
					return nil
				}).ExpectHSet(dataKey, userInvitationData).SetVal(1)
				r.ExpectExpire(dataKey, feature.ExpireDuration()).SetVal(true)
			},
			wantErr:  false,
			isResend: false,
		},
		{
			name: "SUCCESS: User invitation with URL for Resend",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 1}, nil)

				r.ExpectHGet(dataKey, constant.UserInvitationLastTokenField).SetVal(lastToken)

				r.ExpectDel(deleteTokenKey).SetVal(1)

				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)

				r.CustomMatch(func(expected, actual []interface{}) error {
					return nil
				}).ExpectHSet(dataKey, userInvitationData).SetVal(1)
				r.ExpectExpire(dataKey, feature.ExpireDuration()).SetVal(true)
			},
			wantErr:  false,
			isResend: true,
		},
		{
			name: "ERROR: Total resend more than maximum retry",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("6")
			},
			wantErr:  true,
			isResend: true,
		},
		{
			name: "ERROR: Got error rate limiter",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 0}, constant.ErrSomeErrorForUnitTest)
			},
			wantErr:  true,
			isResend: true,
		},
		{
			name: "ERROR: Rate limit is not allowed to access",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 0}, nil)
			},
			wantErr:  true,
			isResend: true,
		},
		{
			name: "ERROR: Got error when scanning last token",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 1}, nil)

				r.ExpectHGet(dataKey, constant.UserInvitationLastTokenField).SetErr(constant.ErrSomeErrorForUnitTest)
			},
			wantErr:  true,
			isResend: true,
		},
		{
			name: "ERROR: Deleting last token",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 1}, nil)

				r.ExpectHGet(dataKey, constant.UserInvitationLastTokenField).SetVal(lastToken)

				r.ExpectDel(deleteTokenKey).SetErr(constant.ErrSomeErrorForUnitTest)

				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)

				r.CustomMatch(func(expected, actual []interface{}) error {
					return nil
				}).ExpectHSet(dataKey, userInvitationData).SetVal(1)
				r.ExpectExpire(dataKey, feature.ExpireDuration()).SetVal(true)
			},
			wantErr:  false,
			isResend: true,
		},
		{
			name: "ERROR: Got error when locking feature",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 1}, nil)

				r.ExpectHGet(dataKey, constant.UserInvitationLastTokenField).SetVal(lastToken)

				r.ExpectDel(deleteTokenKey).SetVal(1)

				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetErr(constant.ErrSomeErrorForUnitTest)
			},
			wantErr:  true,
			isResend: true,
		},
		{
			name: "ERROR: When set user invitation data",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 1}, nil)

				r.ExpectHGet(dataKey, constant.UserInvitationLastTokenField).SetVal(lastToken)

				r.ExpectDel(deleteTokenKey).SetVal(1)

				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)

				r.CustomMatch(func(expected, actual []interface{}) error {
					return nil
				}).ExpectHSet(dataKey, userInvitationData).SetErr(constant.ErrSomeErrorForUnitTest)
			},
			wantErr:  true,
			isResend: true,
		},
		{
			name: "ERROR: When set expire invitation",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 1}, nil)

				r.ExpectHGet(dataKey, constant.UserInvitationLastTokenField).SetVal(lastToken)

				r.ExpectDel(deleteTokenKey).SetVal(1)

				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)

				r.CustomMatch(func(expected, actual []interface{}) error {
					return nil
				}).ExpectHSet(dataKey, userInvitationData).SetVal(1)
				r.ExpectExpire(dataKey, feature.ExpireDuration()).SetErr(constant.ErrSomeErrorForUnitTest)
			},
			wantErr:  true,
			isResend: true,
		},
		{
			name: "ERROR: GenerateInvitationURL",
			mocksSetup: func(r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(false)
			},
			wantErr:  true,
			isResend: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
				MerchantPortalConfig: config.MerchantPortalConfig{
					DashboardGuideURL: "https://example.com",
					LogoURL:           "https://example.com",
					UserInvitationURL: "https://example.com",
				},
			}

			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					UserKey: "testing",
				},
			}

			db, redisClientMock := redismock.NewClientMock()
			redisMock := redisExt.WrapRedisClient(db, nil)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			paperCommMock := repositoryMocks.NewIPaperCommunicationRepository(t)
			limiterMock := redisPkgMocks.NewILimiter(t)

			tc.mocksSetup(redisClientMock, limiterMock)

			trxSvc := New(
				cfg, secret, loggerMock, nil, nil,
				WithRedisClient(redisMock), WithLimiter(limiterMock), WithRabbitMQClient(rmq),
			)

			ctx := context.Background()
			err := trxSvc.SendGeneratedInvitationURL(ctx, &userModel.SendGeneratedInvitationRequest{
				Inviter:      "Inviter",
				Email:        "email@example.id",
				MerchantName: "Merchant Name",
				IsResend:     tc.isResend,
			})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			paperCommMock.AssertExpectations(t)
		})
	}
}
