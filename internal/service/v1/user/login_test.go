package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	rabbitMqPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const errIncorrectCred = "incorrect email or password."

func TestUserServiceLogin(t *testing.T) {
	attempts := &redis.StringCmd{}

	expectedUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Status:     constant.UserStatusActive,
		Name:       "test",
		Password:   "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1",
		MerchantId: "merchant-id",
		CreatedAt:  time.Now(),
	}

	testCases := []struct {
		name         string
		email        string
		password     string
		jwtSecret    string
		attempts     *redis.StringCmd
		inputUser    *userModel.User
		expectedUser *userModel.User
		expectedErr  string
		mockSetup    func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt)
		wantErr      bool
	}{
		{
			name:         "SUCCESS: successfully login",
			email:        "test@gmail.com",
			password:     "pass",
			jwtSecret:    "testing",
			attempts:     attempts,
			expectedUser: expectedUser,
			expectedErr:  "error find user",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userLoggedInDeviceSvc.On(
					"Validate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.BoolMockType(),
				).Return(nil)

				jwtMock.On(
					"GenerateAccessToken",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return(mock.Anything, nil)
				jwtMock.On(
					"GenerateRefreshToken",
					mock.Anything,
					constant.PtrUserMockType(),
					constant.TimeMockType(),
				).Return(mock.Anything, nil)
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedUser, nil)
				trxRepo.On(
					"UpdateRefreshToken",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
				trxRepo.On(
					"Update",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return(nil)
				redis.On(
					"Set",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(nil)
				redis.On(
					"Del",
					mock.Anything,
					constant.StringMockType(),
				).Return(nil)
				jwtMock.On(
					"TerminateTokenOtherDevices",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "ERROR: failed to generate access token",
			email:        "test@gmail.com",
			password:     "pass",
			jwtSecret:    "testing",
			attempts:     attempts,
			expectedUser: expectedUser,
			expectedErr:  "ERROR_INTERNAL",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userLoggedInDeviceSvc.On(
					"Validate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.BoolMockType(),
				).Return(nil)
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedUser, nil)
				jwtMock.On(
					"GenerateAccessToken",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return("", errors.New("failed to generate access token"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: failed to generate refresh token",
			email:        "test@gmail.com",
			password:     "pass",
			jwtSecret:    "testing",
			attempts:     attempts,
			expectedUser: expectedUser,
			expectedErr:  "ERROR_INTERNAL",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userLoggedInDeviceSvc.On(
					"Validate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.BoolMockType(),
				).Return(nil)
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedUser, nil)
				jwtMock.On(
					"GenerateAccessToken",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return(mock.Anything, nil)
				jwtMock.On(
					"GenerateRefreshToken",
					mock.Anything,
					constant.PtrUserMockType(),
					constant.TimeMockType(),
				).Return("", errors.New("failed to generate refresh token"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: find user by email",
			email:        "test@gmail.com",
			password:     "test",
			jwtSecret:    "testing",
			expectedUser: nil,
			expectedErr:  "ERROR_DATABASE",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(nil, errors.New("failed to find user by email"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: user not found",
			email:        "test@gmail.com",
			password:     "test",
			jwtSecret:    "testing",
			expectedUser: nil,
			expectedErr:  "ERROR_NOT_FOUND",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:         "ERROR: failed update user data",
			email:        "test@gmail.com",
			password:     "pass",
			jwtSecret:    "testing",
			attempts:     attempts,
			expectedUser: expectedUser,
			expectedErr:  "failed to update user",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userLoggedInDeviceSvc.On(
					"Validate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.BoolMockType(),
				).Return(nil)
				jwtMock.On(
					"GenerateAccessToken",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return(mock.Anything, nil)
				jwtMock.On(
					"GenerateRefreshToken",
					mock.Anything,
					constant.PtrUserMockType(),
					constant.TimeMockType(),
				).Return(mock.Anything, nil)
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedUser, nil)
				trxRepo.On(
					"UpdateRefreshToken",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
				trxRepo.On(
					"Update",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return(fmt.Errorf("failed to update user"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: invalid_password",
			email:        "test@gmail.com",
			password:     "test",
			jwtSecret:    "testing",
			expectedUser: expectedUser,
			expectedErr:  errIncorrectCred,
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedUser, nil)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
			},
			wantErr: true,
		},
		{
			name:         "ERROR: invalid jwt secret",
			email:        "test@gmail.com",
			password:     "pass",
			jwtSecret:    "",
			expectedUser: expectedUser,
			expectedErr:  "ERROR_DATABASE",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userLoggedInDeviceSvc.On(
					"Validate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.BoolMockType(),
				).Return(nil)
				jwtMock.On(
					"GenerateAccessToken",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return(mock.Anything, nil)
				jwtMock.On(
					"GenerateRefreshToken",
					mock.Anything,
					constant.PtrUserMockType(),
					constant.TimeMockType(),
				).Return(mock.Anything, nil)
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedUser, nil)
				trxRepo.On(
					"UpdateRefreshToken",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("error update refresh token"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: update refresh token",
			email:        "test@gmail.com",
			password:     "pass",
			jwtSecret:    "testing",
			expectedUser: expectedUser,
			expectedErr:  "ERROR_DATABASE",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userLoggedInDeviceSvc.On(
					"Validate", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), constant.BoolMockType(),
				).Return(nil)
				jwtMock.On(
					"GenerateAccessToken",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return(mock.Anything, nil)
				jwtMock.On(
					"GenerateRefreshToken",
					mock.Anything,
					constant.PtrUserMockType(),
					constant.TimeMockType(),
				).Return(mock.Anything, nil)
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedUser, nil)
				trxRepo.On(
					"UpdateRefreshToken",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("error update refresh token"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: reach max attempts",
			email:        "test@gmail.com",
			password:     "pass",
			jwtSecret:    "testing",
			attempts:     attempts,
			expectedUser: expectedUser,
			expectedErr:  "user is blocked",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				attempts.SetVal("5")

				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedUser, nil)
				trxRepo.On(
					"Update",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return(nil)
				redis.On(
					"Del",
					mock.Anything,
					constant.StringMockType(),
				).Return(nil)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
			},
			wantErr: true,
		},
		{
			name:         "ERROR: update_user",
			email:        "test@gmail.com",
			password:     "pass",
			jwtSecret:    "testing",
			attempts:     attempts,
			expectedUser: expectedUser,
			expectedErr:  "ERROR_DATABASE",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)

				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(&userModel.User{
					UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
					Email:      "test@gmail.com",
					Status:     constant.UserStatusActive,
					Name:       "test",
					Password:   "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1",
					MerchantId: "merchant-id",
					CreatedAt:  time.Now(),
				}, nil)

				trxRepo.On(
					"Update",
					mock.Anything,
					constant.PtrUserMockType(),
				).Return(errors.New("error update user"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: user already blocked",
			email:        "test@gmail.com",
			password:     "pass",
			jwtSecret:    "testing",
			attempts:     attempts,
			expectedUser: expectedUser,
			expectedErr:  "ERROR_UNAUTHORIZED",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				attempts.SetVal("5")

				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)

				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedUser, nil)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
			},
			wantErr: true,
		},
		{
			name:      "ERROR: user is deactivated",
			email:     "test@gmail.com",
			password:  "pass",
			jwtSecret: "testing",
			attempts:  attempts,
			expectedUser: &userModel.User{
				UUID:          "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
				Email:         "test@gmail.com",
				Name:          "test",
				Password:      "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1",
				MerchantId:    "merchant-id",
				DeactivatedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
				Blocked:       sql.NullTime{Time: time.Now().Add(-time.Hour * 24), Valid: true},
				CreatedAt:     time.Now(),
			},
			expectedErr: "user is deactivated",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(&userModel.User{
					UUID:          "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
					Email:         "test@gmail.com",
					Name:          "test",
					Password:      "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1",
					MerchantId:    "merchant-id",
					DeactivatedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
					Blocked:       sql.NullTime{Time: time.Now().Add(-time.Hour * 24), Valid: true},
					CreatedAt:     time.Now(),
				}, nil)

			},
			wantErr: true,
		},
		{
			name:      "ERROR: user has not completed onboarding",
			email:     "test@gmail.com", // NOSONAR
			password:  "pass",           // NOSONAR
			jwtSecret: "testing",        // NOSONAR
			attempts:  attempts,
			expectedUser: &userModel.User{
				UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",                             // NOSONAR
				Email:      "test@gmail.com",                                                   // NOSONAR
				Name:       "test",                                                             // NOSONAR
				Password:   "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1", // NOSONAR
				MerchantId: "merchant-id",                                                      // NOSONAR
				CreatedAt:  time.Now().UTC(),                                                   // NOSONAR
				Status:     constant.UserStatusInvited,
			},
			expectedErr: constant.ErrUserInvitedStatus.Error(),
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt, userLoggedInDeviceSvc *serviceMocks.IUserLoggedInDeviceService, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				redis.On(
					"CustomIncr",
					mock.Anything,
					constant.StringMockType(),
					constant.DurationMockType(),
				).Return(attempts)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					constant.StringMockType(),
				).Return(&userModel.User{
					UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",                             // NOSONAR
					Email:      "test@gmail.com",                                                   // NOSONAR
					Name:       "test",                                                             // NOSONAR
					Password:   "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1", // NOSONAR
					MerchantId: "merchant-id",                                                      // NOSONAR
					CreatedAt:  time.Now().UTC(),                                                   // NOSONAR
					Status:     constant.UserStatusInvited,
				}, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}
			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					UserKey: "testing",
				},
			}

			userMock := mockUser.NewIUserRepository(t)
			redisMock := mockRedis.NewIRedisExt(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			passHistoryMock := mockPh.NewIPasswordHistoriesRepository(t)
			jwtMock := mockJWT.NewIJwt(t)
			deviceLoggedInDeviceSvc := serviceMocks.NewIUserLoggedInDeviceService(t)
			rabbitMqMock := rabbitMqPkgMock.NewRabbitMQExt(t)

			tc.mockSetup(userMock, redisMock, jwtMock, deviceLoggedInDeviceSvc, rabbitMqMock)

			ctx := context.WithValue(context.Background(), constant.CtxUserAgentKey, "testing")
			ctx = context.WithValue(ctx, constant.CtxUserDeviceIdentifierKey, "Device-Identifier")

			if tc.name == "ERROR: update_user" {
				// reset block time
				expectedUser.Blocked = sql.NullTime{Time: time.Now().Add(-time.Hour * 24), Valid: true}
			}

			svc := New(
				cfg, secret, loggerMock, userMock, passHistoryMock, WithJWT(jwtMock), WithRedisClient(redisMock),
				WithUserLoggedInDeviceService(deviceLoggedInDeviceSvc), WithRabbitMQClient(rabbitMqMock),
			)
			user, token, err := svc.Login(ctx, &userModel.UserLoginRequest{
				Email:      tc.email,
				Password:   tc.password,
				IsRemember: true,
			})

			if tc.name == "ERROR: invalid_password" && user.EncryptPassword(tc.password) != expectedUser.Password {
				tc.wantErr = true
			}

			if !tc.wantErr {
				assert.NoError(t, err)
				require.NotEmpty(t, user)
				require.NotEmpty(t, token)
			} else {
				require.Error(t, err)
				require.Empty(t, token)

				if !errors.Is(err, constant.ErrBlockedTooManyAttempts) {
					require.Empty(t, user)
				}
			}

			userMock.AssertExpectations(t)
			rabbitMqMock.AssertExpectations(t)
		})
	}
}

var cfg *config.Config

func TestMain(m *testing.M) {
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("/%d-otp-config.yaml", time.Now().Unix()))

	_ = os.WriteFile(tmpFile, []byte(`
MERCHANT_PORTAL:
    IS_NEW_USER_INVITATION_FLOW: true
`), 0777)
	defer os.Remove(tmpFile)

	cfg, _, _ = config.LoadConfig(tmpFile, tmpFile)

	m.Run()
}

func TestUserServiceLoginWithOTP(pt *testing.T) {
	db, clientMock := redismock.NewClientMock()

	logMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	otpSvcMock := serviceMocks.NewIOTP(pt)
	userRepoMock := mockUser.NewIUserRepository(pt)
	jwtMock := mockJWT.NewIJwt(pt)
	rabbitMqMock := rabbitMqPkgMock.NewRabbitMQExt(pt)

	const authAttempKey = "backend-portal:login-attempt:email@example.id"

	mockInvalidValue := []func(r redismock.ClientMock){
		func(r redismock.ClientMock) {
			r.ExpectGet(authAttempKey).SetVal("1")
			r.ExpectIncr(authAttempKey).SetVal(2)
		},
	}
	dataUser := &userModel.User{
		UUID:     "unique-id-1",
		Email:    "email-1@example.id",
		Password: "75768fd714a0fc56a415b8b427b3f09704263f7e438fbfdbea880caae2b13307",
		Status:   constant.UserStatusActive,
	}

	dataUserValid := &userModel.User{
		UUID:     "unique-id-2",
		Email:    "email-2@example.id",
		Password: "75768fd714a0fc56a415b8b427b3f09704263f7e438fbfdbea880caae2b13307",
		Status:   constant.UserStatusActive,
	}

	deviceLoggedInDeviceSvc := serviceMocks.NewIUserLoggedInDeviceService(pt)
	service := New(
		cfg, nil, logMock, userRepoMock, nil,
		WithJWT(jwtMock), WithRedisClient(redisExt.WrapRedisClient(db, nil)), WithOTPService(otpSvcMock), WithUserLoggedInDeviceService(deviceLoggedInDeviceSvc), WithRabbitMQClient(rabbitMqMock),
	)

	tests := []struct {
		name      string
		email     string
		password  string
		mockSetup func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt)
		wantErr   string
		wantToken string
	}{
		{
			name:  "ERROR:Find user by email",
			email: "email@example.id",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(nil, errors.New("invalid session"))
			},
			wantErr: "invalid session",
		},
		{
			name:  "ERROR:Email not found",
			email: "email@example.id",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: errIncorrectCred,
		},
		{
			name:  "ERROR:User is deactivated",
			email: "email@example.id",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(&userModel.User{DeactivatedAt: sql.NullTime{Valid: true}}, nil)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
			},
			wantErr: "user is deactivated",
		},
		{
			name:  "ERROR:User has not completed onboarding",
			email: "email@example.id",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(&userModel.User{Status: constant.UserStatusInvited}, nil)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
			},
			wantErr: constant.ErrUserInvitedStatus.Error(),
		},
		{
			name:  "ERROR:User is not activated",
			email: "email@example.id",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(&userModel.User{Status: ""}, nil)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
			},
			wantErr: "user is not activated",
		},
		{
			name:  "ERROR:User is blocked",
			email: "email@example.id",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(
					&userModel.User{
						Blocked: sql.NullTime{Time: time.Now().UTC().Add(time.Hour)}, Status: constant.UserStatusActive,
					}, nil)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
			},
			wantErr: "user has been blocked",
		},
		{
			name:  "ERROR:Blocked user/Invalid update",
			email: "email@example.id",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				r.ExpectGet(authAttempKey).SetVal("3")
				r.ExpectIncr(authAttempKey).SetVal(4)
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(dataUser, nil)
				u.On(
					"Update", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.PtrUserMockType(),
				).Once().Return(errors.New("invalid update session"))
			},
			wantErr: "invalid update session",
		},
		{
			name:  "ERROR:Blocked user/Success",
			email: "email@example.id",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				r.ExpectGet(authAttempKey).SetVal("3")
				r.ExpectIncr(authAttempKey).SetVal(4)
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(
					&userModel.User{
						UUID:     "unique-id",
						Email:    "email@example.id",
						Password: "75768fd714a0fc56a415b8b427b3f09704263f7e438fbfdbea880caae2b13307",
						Status:   constant.UserStatusActive,
					}, nil)
				u.On(
					"Update", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.PtrUserMockType(),
				).Once().Return(nil)
				j.
					On(
						"RemoveIterateTokenFromRedis",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						constant.StringMockType()).
					Return(nil)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
			},
			wantErr: "user is blocked, too many login attempts",
		},
		{
			name:  "ERROR:Password incorrect",
			email: "email@example.id",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(dataUserValid, nil)
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
			},
			wantErr: errIncorrectCred,
		},
		{
			name:     "ERROR:Generate and send OTP",
			email:    "email@example.id",
			password: "Qwerty123!@#$",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(dataUserValid, nil)
				otpSvcMock.On(
					"GenerateOTPCode", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Once().Return("", errors.New("failed to generate OTP code"))
			},
			wantErr: "failed to generate OTP code",
		},
		{
			name:     "SUCCESS:2FA with OTP",
			email:    "email@example.id",
			password: "Qwerty123!@#$",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(dataUserValid, nil)
				otpSvcMock.On(
					"GenerateOTPCode", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Once().Return(token, nil)
			},
			wantToken: token,
		},
		{
			name:     "SUCCESS:2FA with TOTP",
			email:    "email@example.id",
			password: "Qwerty123!@#$",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				dataUserValid.TOTPStatus = constant.TOTPStatusActive

				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(dataUserValid, nil)
				otpSvcMock.On("GenerateTOTPVerifyToken", mock.Anything, mock.Anything).Once().Return(token, nil)
			},
			wantToken: token,
		},
		{
			name:     "SUCCESS:2FA with preferred OTP method",
			email:    "email@example.id",
			password: "Qwerty123!@#$",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				userWithPreferredOTP := &userModel.User{
					UUID:               "unique-id-3",
					Email:              "email-3@example.id",
					Password:           "75768fd714a0fc56a415b8b427b3f09704263f7e438fbfdbea880caae2b13307",
					Status:             constant.UserStatusActive,
					TOTPStatus:         constant.TOTPStatusActive,
					Preferred2FAMethod: string(constant.TwoFactorAuthMethodOTP),
				}

				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(userWithPreferredOTP, nil)
				otpSvcMock.On(
					"GenerateOTPCode", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Once().Return(token, nil)
			},
			wantToken: token,
		},
		{
			name:     "SUCCESS:2FA with preferred TOTP method",
			email:    "email@example.id",
			password: "Qwerty123!@#$",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				userWithPreferredTOTP := &userModel.User{
					UUID:               "unique-id-4",
					Email:              "email-4@example.id",
					Password:           "75768fd714a0fc56a415b8b427b3f09704263f7e438fbfdbea880caae2b13307",
					Status:             constant.UserStatusActive,
					TOTPStatus:         constant.TOTPStatusActive,
					Preferred2FAMethod: string(constant.TwoFactorAuthMethodTOTP),
				}

				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(userWithPreferredTOTP, nil)
				otpSvcMock.On("GenerateTOTPVerifyToken", mock.Anything, mock.Anything).Once().Return(token, nil)
			},
			wantToken: token,
		},
		{
			name:     "ERROR:Preferred TOTP but not active",
			email:    "email@example.id",
			password: "Qwerty123!@#$",
			mockSetup: func(u *mockUser.IUserRepository, r redismock.ClientMock, j *mockJWT.IJwt, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				for _, f := range mockInvalidValue {
					f(r)
				}
				userWithPreferredTOTP := &userModel.User{
					UUID:               "unique-id-5",
					Email:              "email-5@example.id",
					Password:           "75768fd714a0fc56a415b8b427b3f09704263f7e438fbfdbea880caae2b13307",
					Status:             constant.UserStatusActive,
					TOTPStatus:         constant.TOTPStatusNotEnrolled,
					Preferred2FAMethod: string(constant.TwoFactorAuthMethodTOTP),
				}

				u.On(
					"FindUserByEmail", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(userWithPreferredTOTP, nil)
			},
			wantErr: constant.ErrTOTPRequiredButNotActive.Error(),
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			test.mockSetup(userRepoMock, clientMock, jwtMock, rabbitMqMock)

			token, err := service.LoginWithOTP(context.Background(), test.email, test.password)
			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)

			} else {
				require.Nil(t, err)
				assert.Equal(t, test.wantToken, token)
			}

			jwtMock.AssertExpectations(t)
			otpSvcMock.AssertExpectations(t)
			userRepoMock.AssertExpectations(t)
			rabbitMqMock.AssertExpectations(t)
		})
	}
}

func TestGenerateTokenFromLogin2FA(pt *testing.T) {
	logMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	db, _ := redismock.NewClientMock()

	jwtMock := mockJWT.NewIJwt(pt)
	userRepoMock := mockUser.NewIUserRepository(pt)
	rabbitMqMock := rabbitMqPkgMock.NewRabbitMQExt(pt)
	rabbitMqMock.On(
		"PublishActivity", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType(constant.MockTypeMapStringStringReference),
	).Return(nil)
	rabbitMqMock.On(
		"Publish", mock.Anything, constant.StringMockType(), constant.PtrStringMockType(), mock.Anything,
	).Return(nil)

	service := New(
		cfg, nil, logMock, userRepoMock, nil,
		WithJWT(jwtMock), WithRedisClient(redisExt.WrapRedisClient(db, nil)), WithRabbitMQClient(rabbitMqMock),
	)

	tests := []struct {
		name      string
		id        string
		mockSetup func(u *mockUser.IUserRepository, j *mockJWT.IJwt)
		wantErr   string
		wantToken string
	}{
		{
			name: "ERROR:Find user by ID",
			mockSetup: func(u *mockUser.IUserRepository, _ *mockJWT.IJwt) {
				u.On(
					"FindUserByID", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(nil, errors.New("invalid session"))
			},
			wantErr: "invalid session",
		},
		{
			name: "ERROR:Email not registered",
			mockSetup: func(u *mockUser.IUserRepository, _ *mockJWT.IJwt) {
				u.On(
					"FindUserByID", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: errIncorrectCred,
		},
		{
			name: "ERROR:Generate access token",
			mockSetup: func(u *mockUser.IUserRepository, j *mockJWT.IJwt) {
				u.On(
					"FindUserByID", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(),
				).Return(&userModel.User{}, nil)
				j.On(
					"GenerateAccessToken", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.PtrUserMockType(),
				).Once().Return("", errors.New("failed to generate token"))
			},
			wantErr: "failed to generate token",
		},
		{
			name: "ERROR:Generate refresh token",
			mockSetup: func(_ *mockUser.IUserRepository, j *mockJWT.IJwt) {
				j.On(
					"GenerateAccessToken", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.PtrUserMockType(),
				).Return(token, nil)
				j.On(
					"GenerateRefreshToken", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.PtrUserMockType(), constant.TimeMockType(),
				).Once().Return("", errors.New("failed to generate refresh token"))
			},
			wantErr: "failed to generate refresh token",
		},
		{
			name: "ERROR:Update refresh token",
			mockSetup: func(u *mockUser.IUserRepository, j *mockJWT.IJwt) {
				j.On(
					"GenerateRefreshToken", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.PtrUserMockType(), constant.TimeMockType(),
				).Return(token, nil)
				u.On(
					"UpdateRefreshToken", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(errors.New("invalid session when update token"))
			},
			wantErr: "invalid session when update token",
		},
		{
			name: "ERROR:Update data user",
			mockSetup: func(u *mockUser.IUserRepository, j *mockJWT.IJwt) {
				u.On(
					"UpdateRefreshToken", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
				u.On(
					"Update", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.PtrUserMockType(),
				).Once().Return(errors.New("invalid session when update data user"))
			},
			wantErr: "invalid session when update data user",
		},
		{
			name: "SUCCESS",
			mockSetup: func(u *mockUser.IUserRepository, j *mockJWT.IJwt) {
				u.On(
					"Update", mock.AnythingOfType(constant.MockTypeValueContextReference), constant.PtrUserMockType(),
				).Return(nil)

				jwtMock.On(
					"TerminateTokenOtherDevices",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantToken: token,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			test.mockSetup(userRepoMock, jwtMock)

			ctx := context.WithValue(context.Background(), constant.CtxUserAgentKey, "Go_Unit-Test")
			ctx = context.WithValue(ctx, constant.CtxUserDeviceIdentifierKey, "Device-Identifier")

			_, token, err := service.GenerateTokenFromLogin2FA(ctx, test.id)
			if test.wantErr == "" {
				require.Nil(t, err)
				assert.Equal(t, test.wantToken, token)
			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
