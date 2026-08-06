package user

import (
	"context"
	"errors"
	"testing"

	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/passwordHistories"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_ChangePassword(t *testing.T) {
	userID := uuid.NewString()
	newPassword := "new-password"
	oldPassword := "old-password"

	validUser := &user.User{
		UUID:     userID,
		Password: "ab600ad41f0ad78d3c8f0b76ee7c9136094d87739e14645c54ef6e9ffe26f9d1",
		Status:   constant.UserStatusInvited,
	}

	validUser2 := &user.User{
		UUID:     userID,
		Password: "ab600ad41f0ad78d3c8f0b76ee7c9136094d87739e14645c54ef6e9ffe26f9d1",
	}

	testCases := []struct {
		name       string
		mocksSetup func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository, rateLimiter *mockService.IRateLimiter)
		wantErr    bool
	}{
		{
			name: "ERROR: FindUserByID",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository, rateLimiter *mockService.IRateLimiter) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Old password not match",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository, rateLimiter *mockService.IRateLimiter) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(&user.User{UUID: uuid.NewString()}, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: FindByPassHashAndUserID",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository, rateLimiter *mockService.IRateLimiter) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(validUser, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				phRepo.On(
					"FindByUserID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*int"),
				).Return([]*passwordHistories.PasswordHistories{}, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Password histories not empty",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository, rateLimiter *mockService.IRateLimiter) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(validUser, nil)

				phRepo.On(
					"FindByUserID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*int"),
				).Return([]*passwordHistories.PasswordHistories{
					{
						PasswordHashes: util.HashString("new-password"),
					},
				}, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: ChangePassword repository",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository, rateLimiter *mockService.IRateLimiter) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(validUser, nil)

				phRepo.On(
					"FindByUserID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*int"),
				).Return([]*passwordHistories.PasswordHistories{}, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				trxRepo.On(
					"ChangePassword",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(false, errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Insert password history",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository, rateLimiter *mockService.IRateLimiter) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(validUser, nil)

				phRepo.On(
					"FindByUserID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*int"),
				).Return([]*passwordHistories.PasswordHistories{}, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				trxRepo.On(
					"ChangePassword",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)

				phRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: change password",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository, rateLimiter *mockService.IRateLimiter) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(validUser, nil)

				phRepo.On(
					"FindByUserID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*int"),
				).Return([]*passwordHistories.PasswordHistories{}, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				trxRepo.On(
					"ChangePassword",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)

				phRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				trxRepo.On(
					"Update",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.User"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: change password service error",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository, rateLimiter *mockService.IRateLimiter) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(validUser2, nil)

				phRepo.On(
					"FindByUserID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*int"),
				).Return([]*passwordHistories.PasswordHistories{}, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				trxRepo.On(
					"ChangePassword",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)

				phRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				trxRepo.On(
					"Update",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.User"),
				).Return(constant.ErrSomeErrorForUnitTest)
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

			userRepo := mockUser.NewIUserRepository(t)
			redisMock := mockRedis.NewIRedisExt(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			jwtMock := mockJWT.NewIJwt(t)
			mockPassHistoryRepo := mockPh.NewIPasswordHistoriesRepository(t)
			mockRateLimiter := mockService.NewIRateLimiter(t)

			tc.mocksSetup(userRepo, mockPassHistoryRepo, mockRateLimiter)

			trxSvc := New(
				cfg, secret, loggerMock, userRepo, mockPassHistoryRepo,
				WithJWT(jwtMock), WithRedisClient(redisMock), WithRateLimiter(mockRateLimiter),
			)

			ctx := context.Background()
			_, err := trxSvc.ChangePassword(ctx, userID, oldPassword, newPassword)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
