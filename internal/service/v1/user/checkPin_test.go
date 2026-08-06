package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	rabbitMqPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckPin(t *testing.T) {
	pin := "6 digit pin"

	userWithPin := &userModel.User{
		UUID: "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		PinHash: sql.NullString{
			String: util.HashString(pin),
			Valid:  true,
		},
	}

	userWithInvalidPin := &userModel.User{
		UUID: "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		PinHash: sql.NullString{
			String: "invalid",
			Valid:  true,
		},
	}

	userWithNotSetPin := &userModel.User{
		UUID: "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		PinHash: sql.NullString{
			String: "invalid",
			Valid:  false,
		},
	}

	testCases := []struct {
		name       string
		pin        string
		mocksSetup func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt)
		wantErr    bool
	}{
		{
			name: "SUCCESS: successfully check pin",
			pin:  pin,
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
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
			name: "ERROR: exceeded pin failed attempt",
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(userWithPin, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: find user by ID",
			pin:  pin,
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
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
			pin:  pin,
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: pin not set yet",
			pin:  pin,
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(userWithNotSetPin, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: error check pin",
			pin:  pin,
			mocksSetup: func(userRepo *mockUser.IUserRepository, rateLimiter *mockService.IRateLimiter, rabbitMqMock *rabbitMqPkgMock.RabbitMQExt) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(userWithInvalidPin, nil)

				rateLimiter.On(
					"RateLimitFailedAttempt",
					mock.Anything,
					mock.Anything,
				).Return(nil)

				// Only in the "ERROR: error check pin" case
				rabbitMqMock.On(
					"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
				).Once().Return(nil)
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
			rabbitMqMock := rabbitMqPkgMock.NewRabbitMQExt(t)

			tc.mocksSetup(userRepo, mockRateLimiter, rabbitMqMock)

			trxSvc := New(
				cfg, secret, loggerMock, userRepo, mockPassHistoryRepo,
				WithRedisClient(redisMock), WithJWT(jwtMock), WithRateLimiter(mockRateLimiter), WithRabbitMQClient(rabbitMqMock),
			)

			ctx := context.Background()
			err := trxSvc.CheckCurrentPin(ctx, uuid.NewString(), tc.pin)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
			rabbitMqMock.AssertExpectations(t)
		})
	}
}
