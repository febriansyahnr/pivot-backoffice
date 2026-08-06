package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func TestUserService_Refresh(t *testing.T) {
	jwtSecret := "5f9135b9535d574aedbe693fa4a006de861335f7a48629d73720c79d8854e369"

	expiredAt := time.Now().Add(constant.LOGIN_EXPIRATION_DURATION)
	rClaims := jwt.RegisteredClaims{
		Subject:   "uuid-uuid-uuid",
		Issuer:    "testing",
		ExpiresAt: jwt.NewNumericDate(expiredAt),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, rClaims)

	rt, _ := refreshToken.SignedString([]byte(jwtSecret))

	expectedUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "pass",
		Password:   "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  time.Now(),
	}

	testCases := []struct {
		name         string
		refreshToken string
		email        string
		jwtSecret    string
		expectedUser *userModel.User
		expectedErr  string
		mockSetup    func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt)
		wantErr      bool
	}{
		{
			name:         "SUCCESS: successfully refresh token",
			refreshToken: rt,
			email:        "test@gmail.com",
			jwtSecret:    jwtSecret,
			expectedUser: expectedUser,
			expectedErr:  "",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt) {
				expectedUser.RefreshToken.String = rt

				jwtMock.On(
					"GenerateAccessToken",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(mock.Anything, nil)
				jwtMock.On(
					"GenerateRefreshToken",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType(constant.MockTypeTime),
				).Return(mock.Anything, nil)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"UpdateRefreshToken",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
				redis.On(
					"Set",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("time.Duration"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "ERROR: failed to generate access token",
			refreshToken: rt,
			email:        "test@gmail.com",
			jwtSecret:    jwtSecret,
			expectedUser: expectedUser,
			expectedErr:  "ERROR_INTERNAL",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt) {
				expectedUser.RefreshToken.String = rt

				jwtMock.On(
					"GenerateRefreshToken",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType(constant.MockTypeTime),
				).Return(mock.Anything, nil)
				jwtMock.On(
					"GenerateAccessToken",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return("", errors.New("error generate access token"))
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"UpdateRefreshToken",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			wantErr: true,
		},
		{
			name:         "ERROR: failed to generate refresh token",
			refreshToken: rt,
			email:        "test@gmail.com",
			jwtSecret:    jwtSecret,
			expectedUser: expectedUser,
			expectedErr:  "ERROR_INTERNAL",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt) {
				expectedUser.RefreshToken.String = rt

				jwtMock.On(
					"GenerateRefreshToken",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType(constant.MockTypeTime),
				).Return("", errors.New("error generate refresh token"))
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
			},
			wantErr: true,
		},
		{
			name:         "ERROR: failed to find user",
			refreshToken: rt,
			email:        "test@gmail.com",
			jwtSecret:    jwtSecret,
			expectedUser: expectedUser,
			expectedErr:  "ERROR_DATABASE",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt) {
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("error when finding user by email"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: failed to update refresh token",
			refreshToken: rt,
			email:        "test@gmail.com",
			jwtSecret:    jwtSecret,
			expectedUser: expectedUser,
			expectedErr:  "ERROR_DATABASE",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt) {
				expectedUser.RefreshToken.String = rt

				jwtMock.On(
					"GenerateRefreshToken",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType(constant.MockTypeTime),
				).Return(mock.Anything, nil)
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
				trxRepo.On(
					"UpdateRefreshToken",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("error update refresh token"))
			},
			wantErr: true,
		},
		{
			name:         "ERROR: refresh token is not the same with user's refresh token",
			refreshToken: "invalid-token",
			email:        "test@gmail.com",
			jwtSecret:    jwtSecret,
			expectedUser: expectedUser,
			expectedErr:  "ERROR_UNAUTHORIZED",
			mockSetup: func(trxRepo *mockUser.IUserRepository, redis *mockRedis.IRedisExt, jwtMock *mockJWT.IJwt) {
				trxRepo.On(
					"FindUserByEmail",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
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
			jwtMock := mockJWT.NewIJwt(t)
			passHistoryMock := mockPh.NewIPasswordHistoriesRepository(t)

			tc.mockSetup(userMock, redisMock, jwtMock)

			ctx := context.WithValue(context.Background(), constant.CtxUserAgentKey, "testing")

			svc := New(
				cfg, secret, loggerMock, userMock, passHistoryMock,
				WithJWT(jwtMock), WithRedisClient(redisMock),
			)
			newRefresh, token, err := svc.Refresh(ctx, "", tc.refreshToken)

			if !tc.wantErr {
				assert.NoError(t, err)
				require.NotEmpty(t, newRefresh)
				require.NotEmpty(t, token)
			} else {
				require.Error(t, err)
				require.Empty(t, newRefresh)
				require.Empty(t, token)
				require.True(t, strings.Contains(err.Error(), tc.expectedErr))
			}

			userMock.AssertExpectations(t)
		})
	}
}
