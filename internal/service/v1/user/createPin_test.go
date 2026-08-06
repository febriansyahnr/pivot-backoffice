package user

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePin(t *testing.T) {
	pin := "132435"

	expectedUser := &userModel.User{
		UUID: "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
	}

	userWithPin := &userModel.User{
		UUID: "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		PinHash: sql.NullString{
			String: util.HashString(pin),
			Valid:  true,
		},
	}

	testCases := []struct {
		name       string
		pin        string
		mocksSetup func(userRepo *mockUser.IUserRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS: successfully create pin",
			pin:  pin,
			mocksSetup: func(userRepo *mockUser.IUserRepository) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)

				userRepo.On(
					"UpdatePin",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: find user by ID",
			pin:  pin,
			mocksSetup: func(userRepo *mockUser.IUserRepository) {
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
			mocksSetup: func(userRepo *mockUser.IUserRepository) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: pin is not empty",
			pin:  pin,
			mocksSetup: func(userRepo *mockUser.IUserRepository) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(userWithPin, nil)

			},
			wantErr: true,
		},
		{
			name: "ERROR: error create pin",
			pin:  pin,
			mocksSetup: func(userRepo *mockUser.IUserRepository) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)

				userRepo.On(
					"UpdatePin",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: invalid pin format",
			pin:  "11111",
			mocksSetup: func(userRepo *mockUser.IUserRepository) {
				userRepo.On(
					"FindUserByID", mock.Anything,
					constant.StringMockType()).Return(expectedUser, nil)
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

			tc.mocksSetup(userRepo)

			trxSvc := New(
				cfg, secret, loggerMock, userRepo, mockPassHistoryRepo, WithJWT(jwtMock), WithRedisClient(redisMock),
			)

			ctx := context.Background()
			err := trxSvc.CreatePin(ctx, uuid.NewString(), tc.pin)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
