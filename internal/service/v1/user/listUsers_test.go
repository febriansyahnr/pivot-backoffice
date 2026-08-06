package user

import (
	"context"
	"database/sql"
	"fmt"
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

func TestUserService_ListUsers(t *testing.T) {
	dummyUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "pass",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  time.Now(),
	}

	expectedUsers := []*userModel.User{dummyUser}

	testCases := []struct {
		name          string
		IsSuccess     bool
		limit         int
		offset        int
		ExpectedError string
		mockSetup     func(mockRepo *mockUser.IUserRepository)
	}{
		{
			name:          "SUCCESS: list users",
			IsSuccess:     true,
			limit:         10,
			offset:        10,
			ExpectedError: "",
			mockSetup: func(mockRepo *mockUser.IUserRepository) {
				mockRepo.On(
					"ListUsers",
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int"),
				).Return(expectedUsers, nil)
			},
		},
		{
			name:          "ERROR: error list users",
			IsSuccess:     false,
			limit:         10,
			offset:        0,
			ExpectedError: "error list users",
			mockSetup: func(mockRepo *mockUser.IUserRepository) {
				mockRepo.On(
					"ListUsers",
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int"),
				).Return(nil, fmt.Errorf("error list users"))
			},
		},
		{
			name:          "ERROR: no users found",
			IsSuccess:     false,
			limit:         10,
			offset:        0,
			ExpectedError: "ERROR_NOT_FOUND",
			mockSetup: func(mockRepo *mockUser.IUserRepository) {
				mockRepo.On(
					"ListUsers",
					mock.Anything,
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int"),
				).Return(nil, nil)
			},
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
			ctx := context.Background()

			tc.mockSetup(userMock)

			svc := New(
				cfg, secret, loggerMock, userMock, passHistoryMock, WithJWT(jwtMock), WithRedisClient(redisMock),
			)

			results, err := svc.ListUsers(ctx, tc.limit, tc.offset)

			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, results)
			} else {
				require.Error(t, err)
				require.Empty(t, results)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			}

			userMock.AssertExpectations(t)
		})
	}
}
