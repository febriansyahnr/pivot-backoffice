package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_GenerateRandomPassword(t *testing.T) {
	validUser := &user.User{
		UUID: uuid.NewString(),
	}

	testCases := []struct {
		name       string
		mocksSetup func(trxRepo *mockUser.IUserRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS: generate password",
			mocksSetup: func(trxRepo *mockUser.IUserRepository) {
				trxRepo.On(
					"ChangePassword",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)

			},
			wantErr: false,
		},
		{
			name: "ERROR: ChangePassword repository",
			mocksSetup: func(trxRepo *mockUser.IUserRepository) {
				trxRepo.On(
					"ChangePassword",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(false, errors.New("some-error"))

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
			_, err := trxSvc.GenerateRandomPassword(ctx, validUser)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
