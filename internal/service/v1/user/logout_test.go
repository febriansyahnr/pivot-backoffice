package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_Logout(t *testing.T) {
	expectedUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1",
		MerchantId: "merchant-id",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour * 24), Valid: true},
		CreatedAt:  time.Now(),
	}

	testCases := []struct {
		name         string
		email        string
		jwtSecret    string
		inputUser    *userModel.User
		expectedUser *userModel.User
		expectedErr  string
		mockSetup    func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt)
		wantErr      bool
	}{
		{
			name:         "SUCCESS: successfully login",
			email:        "test@gmail.com",
			jwtSecret:    "testing",
			expectedUser: expectedUser,
			expectedErr:  "error find user",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt) {
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				redis.On(
					"Del",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil)
				redis.On(
					"Del",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "ERROR: find user by email",
			email:        "test@gmail.com",
			jwtSecret:    "testing",
			expectedUser: nil,
			expectedErr:  "ERROR_DATABASE",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt) {
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("error find user"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: user not found",
			email:        "test@gmail.com",
			jwtSecret:    "testing",
			expectedUser: nil,
			expectedErr:  "ERROR_NOT_FOUND",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt) {
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
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

			userMock := mockUser.NewIUserRepository(t)
			redisMock := mockRedis.NewIRedisExt(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			passHistoryMock := mockPh.NewIPasswordHistoriesRepository(t)
			jwtMock := mockJWT.NewIJwt(t)

			tc.mockSetup(userMock, redisMock)

			ctx := context.WithValue(context.Background(), constant.CtxUserAgentKey, "testing")

			svc := New(
				cfg, secret, loggerMock, userMock, passHistoryMock,
				WithJWT(jwtMock), WithRedisClient(redisMock),
			)
			err := svc.Logout(ctx, tc.email)

			if !tc.wantErr {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tc.expectedErr))
			}

			userMock.AssertExpectations(t)
		})
	}
}
