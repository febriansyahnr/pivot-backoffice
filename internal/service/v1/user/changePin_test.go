package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestChangePin(t *testing.T) {
	pin := "132435"
	newPin := "243546"

	expectedUser := &userModel.User{
		UUID: "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		PinHash: sql.NullString{
			Valid:  true,
			String: util.HashString(pin),
		},
	}

	invalidPinUser := &userModel.User{
		UUID: "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		PinHash: sql.NullString{
			Valid:  true,
			String: util.HashString("invalid pin"),
		},
	}

	testCases := []struct {
		name       string
		mocksSetup func(userRepo *mockUser.IUserRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS: successfully change pin",
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
			name: "ERROR: invalid pin",
			mocksSetup: func(userRepo *mockUser.IUserRepository) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(invalidPinUser, nil)

			},
			wantErr: true,
		},
		{
			name: "ERROR: pin not created yet",
			mocksSetup: func(userRepo *mockUser.IUserRepository) {
				userRepo.On(
					"FindUserByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&userModel.User{}, nil)

			},
			wantErr: true,
		},
		{
			name: "ERROR: error change pin",
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
			err := trxSvc.ChangePin(ctx, uuid.NewString(), pin, newPin)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestResetPIN(pt *testing.T) {
	userMock := mockUser.NewIUserRepository(pt)

	service := New(&config.Config{}, nil, nil, userMock, nil)

	tests := []struct {
		name      string
		mockSetup func(u *mockUser.IUserRepository)
		wantErr   string
	}{
		{
			name: "ERROR:Find user by ID",
			mockSetup: func(u *mockUser.IUserRepository) {
				u.On(
					"FindUserByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, pkgErrs.New(response.HttpErrDatabase, errors.New("invalid session")))
			},
			wantErr: "invalid session",
		},
		{
			name: "ERROR:User not found",
			mockSetup: func(u *mockUser.IUserRepository) {
				u.On(
					"FindUserByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrUserNotFound))
			},
			wantErr: "user not found",
		},
		{
			name: "ERROR:Update PIN",
			mockSetup: func(u *mockUser.IUserRepository) {
				u.On(
					"FindUserByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(&userModel.User{UUID: "unique-id"}, nil)
				u.On(
					"UpdatePin", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(errors.New("invalid arguments"))
			},
			wantErr: "invalid argument",
		},
		{
			name: "SUCCESS",
			mockSetup: func(u *mockUser.IUserRepository) {
				u.On(
					"UpdatePin", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			test.mockSetup(userMock)

			err := service.ResetPIN(context.Background(), "unique-id", "132435")
			if test.wantErr == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
