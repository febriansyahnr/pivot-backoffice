package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	createdUser := &userModel.User{
		UUID:       uuid.New().String(),
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "test",
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "test",
		CreatedAt:  time.Now(),
	}

	testCases := []struct {
		name       string
		input      *userModel.User
		mocksSetup func(trxRepo *mockUser.IUserRepository)
		wantErr    bool
	}{
		{
			name:  "SUCCESS: successfully create user",
			input: createdUser,
			mocksSetup: func(trxRepo *mockUser.IUserRepository) {
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: error create user",
			input: createdUser,
			mocksSetup: func(trxRepo *mockUser.IUserRepository) {
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*user.User"),
				).Return(errors.New("insert error"))
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
				cfg, secret, loggerMock, userRepo, mockPassHistoryRepo,
				WithRedisClient(redisMock), WithJWT(jwtMock),
			)

			ctx := context.Background()
			err := trxSvc.Create(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
