package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	rabbitMqPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestActivateUser(t *testing.T) {
	validRequest := &userModel.ActivateUserRequest{
		Email:    "test@mail.com",
		Password: util.HashString("pass"),
		PIN:      util.HashString("pin"),
		Token:    "1234",
	}

	validUser := &userModel.User{
		UUID:   uuid.NewString(),
		Status: constant.UserStatusInvited,
	}

	cfg := &config.Config{
		ServiceName: "testing",
	}

	secret := &config.Secret{
		JWTSignatureKey: config.JWTSignatureKey{
			UserKey: "testing",
		},
	}

	userRepo := mockUser.NewIUserRepository(t)
	loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	jwtMock := mockJWT.NewIJwt(t)

	db, clientMock := redismock.NewClientMock()
	redisMock := redisExt.WrapRedisClient(db, nil)

	rabbitMqMock := rabbitMqPkgMock.NewRabbitMQExt(t)
	rabbitMqMock.On(
		"PublishActivity", constant.ValueCtxMockType(), constant.PtrStringMockType(), constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.MapStrValStringMockType(),
	).Once().Return(nil)
	rabbitMqMock.On(
		"Publish", mock.Anything, constant.StringMockType(), constant.PtrStringMockType(), mock.Anything,
	).Return(nil)

	trxSvc := New(
		cfg, secret, loggerMock, userRepo, nil,
		WithJWT(jwtMock), WithRedisClient(redisMock), WithRabbitMQClient(rabbitMqMock),
	)

	testCases := []struct {
		name       string
		request    *userModel.ActivateUserRequest
		mocksSetup func(r redismock.ClientMock, userRepo *mockUser.IUserRepository, j *mockJWT.IJwt)
		wantErr    bool
	}{
		{
			name:    "SUCCESS",
			request: validRequest,
			mocksSetup: func(r redismock.ClientMock, userRepo *mockUser.IUserRepository, j *mockJWT.IJwt) {
				userRepo.On(
					"FindUserByEmail",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Once().Return(validUser, nil)

				userRepo.On(
					"Update",
					constant.ValueCtxMockType(),
					constant.PtrUserMockType(),
				).Return(nil)

				userRepo.On(
					"FindUserByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&userModel.User{}, nil)

				j.On(
					"GenerateAccessToken", constant.ValueCtxMockType(), constant.PtrUserMockType(),
				).Return(token, nil)

				j.On(
					"GenerateRefreshToken", constant.ValueCtxMockType(), constant.PtrUserMockType(), constant.TimeMockType(),
				).Return(token, nil)

				userRepo.On(
					"UpdateRefreshToken", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)

				r.ExpectKeys(fmt.Sprintf("backend-portal:users::user-invitation:token:%s", validRequest.Token)).SetVal([]string{"key-1"})
				r.ExpectDel(fmt.Sprintf("backend-portal:users::user-invitation:token:%s", validRequest.Token)).SetVal(1)

				jwtMock.On(
					"TerminateTokenOtherDevices",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: User not found",
			request: validRequest,
			mocksSetup: func(r redismock.ClientMock, userRepo *mockUser.IUserRepository, j *mockJWT.IJwt) {
				userRepo.On(
					"FindUserByEmail",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Service FindUserByEmail error",
			request: validRequest,
			mocksSetup: func(r redismock.ClientMock, userRepo *mockUser.IUserRepository, j *mockJWT.IJwt) {
				userRepo.On(
					"FindUserByEmail",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Service Update error",
			request: validRequest,
			mocksSetup: func(r redismock.ClientMock, userRepo *mockUser.IUserRepository, j *mockJWT.IJwt) {
				userRepo.On(
					"FindUserByEmail",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&userModel.User{
					UUID:   uuid.NewString(),
					Status: constant.UserStatusInvited,
				}, nil)

				userRepo.On(
					"Update",
					constant.ValueCtxMockType(),
					constant.PtrUserMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Already activated",
			request: validRequest,
			mocksSetup: func(r redismock.ClientMock, userRepo *mockUser.IUserRepository, j *mockJWT.IJwt) {
				userRepo.On(
					"FindUserByEmail",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&userModel.User{Status: constant.UserStatusActive}, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			tc.mocksSetup(clientMock, userRepo, jwtMock)
			ctx := context.WithValue(context.Background(), constant.CtxUserAgentKey, "user-agent")
			_, err := trxSvc.ActivateUser(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
