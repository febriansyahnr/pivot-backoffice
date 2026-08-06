package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
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

func TestFindUserByID(t *testing.T) {
	expectedUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "pass",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  time.Now(),
	}

	testCases := []struct {
		Name           string
		IsSuccess      bool
		UserID         string
		ExpectedResult *userModel.UserResponse
		ExpectedError  string
		MockSetup      func(mockRepo *mockUser.IUserRepository)
	}{
		{
			Name:           "SUCCESS: find user by id",
			IsSuccess:      true,
			UserID:         "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
			ExpectedResult: expectedUser.ToResponse(),
			MockSetup: func(mockRepo *mockUser.IUserRepository) {
				mockRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedUser, nil)
			},
		},
		{
			Name:          "ERROR: user not found",
			IsSuccess:     false,
			UserID:        "not-found",
			ExpectedError: "ERROR_NOT_FOUND",
			MockSetup: func(mockRepo *mockUser.IUserRepository) {
				mockRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
		},
		{
			Name:          "ERROR: error find user",
			IsSuccess:     false,
			UserID:        "user-error",
			ExpectedError: "ERROR_DATABASE",
			MockSetup: func(mockRepo *mockUser.IUserRepository) {
				mockRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("error when finding user by id"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
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
			ctx := context.Background()

			tc.MockSetup(userMock)

			userService := New(
				cfg, secret, loggerMock, userMock, passHistoryMock,
				WithJWT(jwtMock), WithRedisClient(redisMock),
			)

			response, err := userService.FindUserByID(ctx, tc.UserID)
			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			} else {
				require.Error(t, err)
				require.Empty(t, response)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			}

			userMock.AssertExpectations(t)
		})
	}
}
