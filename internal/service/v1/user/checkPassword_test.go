package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckCurrentPassword(t *testing.T) {
	password := "dummy-password"

	userWithPin := &userModel.User{
		UUID:     "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Password: util.HashString(password),
	}

	userWithInvalidPassword := &userModel.User{
		UUID:     "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Password: "invalid-password",
	}

	testCases := []struct {
		name       string
		pin        string
		mocksSetup func(userRepo *mockUser.IUserRepository, rateLimitMock *mockService.IRateLimiter)
		wantErr    bool
	}{
		{
			name: "SUCCESS: successfully check password",
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(userWithPin, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: find user by ID",
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: user not found",
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: error check password",
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(userWithInvalidPassword, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)

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

			tc.mocksSetup(userRepo, mockRateLimiter)

			trxSvc := New(
				cfg, secret, loggerMock, userRepo, mockPassHistoryRepo,
				WithRedisClient(redisMock), WithJWT(jwtMock), WithRateLimiter(mockRateLimiter),
			)

			ctx := context.Background()
			err := trxSvc.CheckCurrentPassword(ctx, uuid.NewString(), password)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
