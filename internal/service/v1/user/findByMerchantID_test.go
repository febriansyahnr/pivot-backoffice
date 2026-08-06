package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
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

func TestUserService_ListUsersByMerchantID(t *testing.T) {
	data := make([]userModel.User, 0)
	data = append(data, userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "pass",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  time.Now(),
	})

	input := &userModel.ListUsersByMerchantIDRequest{}

	expectedResponse := commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		name           string
		input          *userModel.ListUsersByMerchantIDRequest
		WantErr        bool
		ExpectedResult *commonModel.PaginationResponse
		ExpectedError  string
		mockSetup      func(mockRepo *mockUser.IUserRepository)
	}{
		{
			name:           "SUCCESS: list users",
			input:          input,
			WantErr:        false,
			ExpectedResult: &expectedResponse,
			mockSetup: func(mockRepo *mockUser.IUserRepository) {
				mockRepo.On(
					"ListUsersByMerchantID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.ListUsersByMerchantIDRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(&expectedResponse, nil)
			},
		},
		{
			name: "ERROR: Failed request validation",
			input: &userModel.ListUsersByMerchantIDRequest{
				SortBy: "names"},
			WantErr:        true,
			ExpectedResult: &expectedResponse,
			mockSetup: func(mockRepo *mockUser.IUserRepository) {
			},
		},
		{
			name:           "ERROR: error list users",
			input:          input,
			WantErr:        true,
			ExpectedResult: &expectedResponse,
			ExpectedError:  "ERROR_DATABASE | error list users",
			mockSetup: func(mockRepo *mockUser.IUserRepository) {
				mockRepo.On(
					"ListUsersByMerchantID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.ListUsersByMerchantIDRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, fmt.Errorf("error list users"))
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
				cfg, secret, loggerMock, userMock, passHistoryMock,
				WithJWT(jwtMock), WithRedisClient(redisMock),
			)

			results, err := svc.ListUsersByMerchantID(ctx, tc.input, 1, 10)

			if tc.WantErr {
				require.Error(t, err)
				require.Empty(t, results)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, results)
			}

			userMock.AssertExpectations(t)
		})
	}
}
