package user

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	loginSuspendKey = "backend-portal:otp-verification:email:user-login:suspend"
)

func TestUserService_UnblockUser(t *testing.T) {
	expectedUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Status:     constant.UserStatusActive,
		Name:       "test",
		Password:   "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1",
		MerchantId: "merchant-id",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour * 24), Valid: true},
		CreatedAt:  time.Now(),
	}

	redisString := &redis.StringCmd{}

	testCases := []struct {
		name              string
		email             string
		expectedUser      *userModel.User
		expectedErr       string
		expectedStringCmd *redis.StringCmd
		mockSetup         func(userRepoMock *mockUser.IUserRepository, redis *mockRedis.IRedisExt, r redismock.ClientMock)
		wantErr           bool
	}{
		{
			name:              "SUCCESS: unblock user successfully",
			email:             expectedUser.Email,
			expectedErr:       "",
			expectedUser:      expectedUser,
			expectedStringCmd: redisString,
			mockSetup: func(userRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, r redismock.ClientMock) {
				userRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)

				r.ClearExpect()
				r.ExpectGet(loginSuspendKey).SetVal(`{"status": false, "retry_after": ""}`)
				r.ExpectDel(loginSuspendKey).SetVal(1)
				r.ExpectDel(fmt.Sprintf("backend-portal:login-attempt:%s", expectedUser.Email)).SetVal(1)

				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(nil)

			},
			wantErr: false,
		},
		{
			name:              "FAILED: failed to get user by email",
			email:             expectedUser.Email,
			expectedErr:       "failed to get user by email",
			expectedUser:      nil,
			expectedStringCmd: redisString,
			mockSetup: func(userRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, r redismock.ClientMock) {
				userRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, fmt.Errorf("failed to get user by email"))

			},
			wantErr: true,
		},
		{
			name:              "FAILED: user not found",
			email:             expectedUser.Email,
			expectedErr:       "user not found",
			expectedUser:      nil,
			expectedStringCmd: redisString,
			mockSetup: func(userRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, r redismock.ClientMock) {
				userRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)

			},
			wantErr: true,
		},
		{
			name:              "FAILED: failed to update user",
			email:             expectedUser.Email,
			expectedErr:       "failed to update user",
			expectedUser:      nil,
			expectedStringCmd: redisString,
			mockSetup: func(userRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, r redismock.ClientMock) {
				userRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)

				r.ClearExpect()
				r.ExpectGet(loginSuspendKey).SetVal(`{"status": false, "retry_after": ""}`)
				r.ExpectDel(loginSuspendKey).SetVal(1)
				r.ExpectDel(fmt.Sprintf("backend-portal:login-attempt:%s", expectedUser.Email)).SetVal(1)

				userRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(fmt.Errorf("failed to update user"))

			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, clientMock := redismock.NewClientMock()
			redisMock := redisExt.WrapRedisClient(db, nil)

			cfg := &config.Config{
				ServiceName: "testing",
			}
			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					UserKey: "testing",
				},
			}

			userMock := mockUser.NewIUserRepository(t)
			redisExtMock := mockRedis.NewIRedisExt(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			passHistoryMock := mockPh.NewIPasswordHistoriesRepository(t)
			jwtMock := mockJWT.NewIJwt(t)

			tc.mockSetup(userMock, redisExtMock, clientMock)

			ctx := context.WithValue(context.Background(), constant.CtxUserAgentKey, "testing")

			svc := New(
				cfg, secret, loggerMock, userMock, passHistoryMock, WithJWT(jwtMock), WithRedisClient(redisMock),
			)
			err := svc.UnblockUser(ctx, tc.email)

			if !tc.wantErr {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
