package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/passwordHistories"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_UserDetail(t *testing.T) {
	validUser := &user.User{
		UUID: uuid.NewString(),
	}

	now := time.Now()
	passwordHistory := []*passwordHistories.PasswordHistories{
		{
			UserID:    validUser.UUID,
			CreatedAt: now,
		},
	}

	testCases := []struct {
		name                     string
		mocksSetup               func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository)
		wantErr                  bool
		wantLastPasswordChange   bool
		expectedLastPasswordTime *time.Time
	}{
		{
			name: "SUCCESS: get user detail without password history",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository) {
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
				).Return(nil, nil)
			},
			wantErr:                false,
			wantLastPasswordChange: false,
		},
		{
			name: "SUCCESS: get user detail with password history",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository) {
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
				).Return(passwordHistory, nil)
			},
			wantErr:                  false,
			wantLastPasswordChange:   true,
			expectedLastPasswordTime: &now,
		},
		{
			name: "SUCCESS: get user detail even when password history fails",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository) {
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
				).Return(nil, errors.New("database error"))
			},
			wantErr:                false,
			wantLastPasswordChange: false,
		},
		{
			name: "ERROR: FindUserByID",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("some-error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: User not found",
			mocksSetup: func(trxRepo *mockUser.IUserRepository, phRepo *mockPh.IPasswordHistoriesRepository) {
				trxRepo.On(
					"FindUserByID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, nil)
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

			tc.mocksSetup(userRepo, mockPassHistoryRepo)

			trxSvc := New(
				cfg, secret, loggerMock, userRepo, mockPassHistoryRepo,
				WithJWT(jwtMock), WithRedisClient(redisMock),
			)

			ctx := context.Background()
			result, err := trxSvc.UserDetail(ctx, validUser.UUID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Check if lastChangePassword is set correctly
				if tc.wantLastPasswordChange {
					assert.NotNil(t, result.LastChangePassword, "LastChangePassword should not be nil")
					if tc.expectedLastPasswordTime != nil {
						assert.Equal(t, tc.expectedLastPasswordTime.Unix(), result.LastChangePassword.Unix(), "LastChangePassword time should match")
					}
				} else {
					assert.Nil(t, result.LastChangePassword, "LastChangePassword should be nil")
				}
			}

			userRepo.AssertExpectations(t)
			mockPassHistoryRepo.AssertExpectations(t)
		})
	}
}
