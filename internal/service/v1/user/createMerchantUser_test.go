package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisPkgMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateMerchantUser(t *testing.T) {
	defLockKey := "backend-portal:users:email@example.id:user-invitation:lock"
	dataKey := "backend-portal:users:email@example.id:user-invitation:data"
	lastToken := "last-token"
	userInvitationData := map[string]interface{}{
		constant.UserInvitationTotalResendField: 0,
		constant.UserInvitationLastTokenField:   lastToken,
	}
	feature := constant.UserIdentifierUserInvitation

	req := &userModel.MerchantUserRequest{
		Email:        "email@example.id",
		Name:         "name",
		Role:         "admin",
		MerchantId:   uuid.NewString(),
		MerchantName: "merchant name",
		Invitation:   true,
	}
	rmq := rmqMock.NewRabbitMQExt(t)
	rmq.On(
		"Publish", constant.ValueCtxMockType(), constant.StringMockType(), mock.Anything, mock.AnythingOfType("[]uint8"),
	).Return(nil)

	testCases := []struct {
		Name      string
		Setup     func(repo *mockRepo.IUserRepository, roleSvc *mocks.IRoleService, userRoleSvc *mocks.IUserRoleService, r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter)
		ExpectErr bool
	}{
		{
			Name: "ERROR: Find user by email",
			Setup: func(repo *mockRepo.IUserRepository, roleSvc *mocks.IRoleService, userRoleSvc *mocks.IUserRoleService, r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				repo.On("FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, errors.New("errors"))
			},
			ExpectErr: true,
		},
		{
			Name: "ERROR: Existing user found",
			Setup: func(repo *mockRepo.IUserRepository, roleSvc *mocks.IRoleService, userRoleSvc *mocks.IUserRoleService, r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				repo.On("FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType()).Return(&userModel.User{UUID: uuid.NewString()}, nil)
			},
			ExpectErr: true,
		},
		{
			Name: "ERROR: Create",
			Setup: func(repo *mockRepo.IUserRepository, roleSvc *mocks.IRoleService, userRoleSvc *mocks.IUserRoleService, r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				repo.On("FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*user.User")).Return(errors.New("errors"))
			},
			ExpectErr: true,
		},
		{
			Name: "ERROR: Find role by slug",
			Setup: func(repo *mockRepo.IUserRepository, roleSvc *mocks.IRoleService, userRoleSvc *mocks.IUserRoleService, r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				repo.On("FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*user.User")).Return(nil)
				roleSvc.On("FindRoleBySlug", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, errors.New("errors"))
			},
			ExpectErr: true,
		},
		{
			Name: "ERROR: Role not found",
			Setup: func(repo *mockRepo.IUserRepository, roleSvc *mocks.IRoleService, userRoleSvc *mocks.IUserRoleService, r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				repo.On("FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*user.User")).Return(nil)
				roleSvc.On("FindRoleBySlug", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
			},
			ExpectErr: true,
		},
		{
			Name: "ERROR: Create user role",
			Setup: func(repo *mockRepo.IUserRepository, roleSvc *mocks.IRoleService, userRoleSvc *mocks.IUserRoleService, r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				repo.On("FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*user.User")).Return(nil)
				roleSvc.On("FindRoleBySlug", constant.ValueCtxMockType(), constant.StringMockType()).Return(&role.Role{UUID: uuid.NewString()}, nil)
				userRoleSvc.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*userRole.UserRole")).Return(errors.New("errors"))
			},
			ExpectErr: true,
		},
		{
			Name: "ERROR: Send invite",
			Setup: func(repo *mockRepo.IUserRepository, roleSvc *mocks.IRoleService, userRoleSvc *mocks.IUserRoleService, r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				repo.On("FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*user.User")).Return(nil)
				roleSvc.On("FindRoleBySlug", constant.ValueCtxMockType(), constant.StringMockType()).Return(&role.Role{UUID: uuid.NewString()}, nil)
				userRoleSvc.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*userRole.UserRole")).Return(nil)
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("6")

			},
			ExpectErr: true,
		},
		{
			Name: "SUCCESS: Create Merchant User",
			Setup: func(repo *mockRepo.IUserRepository, roleSvc *mocks.IRoleService, userRoleSvc *mocks.IUserRoleService, r redismock.ClientMock, limiterMock *redisPkgMocks.ILimiter) {
				repo.On("FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
				repo.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*user.User")).Return(nil)
				roleSvc.On("FindRoleBySlug", constant.ValueCtxMockType(), constant.StringMockType()).Return(&role.Role{UUID: uuid.NewString()}, nil)
				userRoleSvc.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*userRole.UserRole")).Return(nil)
				r.ExpectHGet(dataKey, constant.UserInvitationTotalResendField).SetVal("0")

				limiterMock.On(
					"Allow", mock.Anything, constant.StringMockType(), mock.AnythingOfType("*redis_rate.Limit"),
				).Return(&redisExt.Result{Allowed: 1}, nil)

				r.ExpectHGet(dataKey, constant.UserInvitationLastTokenField).SetVal(lastToken)
				r.ExpectSetNX(defLockKey, true, 10*time.Second).SetVal(true)

				r.CustomMatch(func(expected, actual []interface{}) error {
					return nil
				}).ExpectHSet(dataKey, userInvitationData).SetVal(1)
				r.ExpectExpire(dataKey, feature.ExpireDuration()).SetVal(true)
			},
			ExpectErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
				MerchantPortalConfig: config.MerchantPortalConfig{
					DashboardGuideURL: "https://example.com",
					LogoURL:           "https://example.com",
					UserInvitationURL: "https://example.com",
				},
			}

			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					UserKey: "testing",
				},
			}

			userRepo := mockRepo.NewIUserRepository(t)
			roleSvc := mocks.NewIRoleService(t)
			userRoleSvc := mocks.NewIUserRoleService(t)
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			db, redisClientMock := redismock.NewClientMock()
			redisMock := redisExt.WrapRedisClient(db, nil)
			limiterMock := redisPkgMocks.NewILimiter(t)

			tc.Setup(userRepo, roleSvc, userRoleSvc, redisClientMock, limiterMock)
			userSvc := New(cfg, secret, logger, userRepo, nil,
				WithUserRoleService(userRoleSvc), WithRoleService(roleSvc), WithRedisClient(redisMock), WithLimiter(limiterMock), WithRabbitMQClient(rmq),
			)
			_, err := userSvc.CreateMerchantUser(context.Background(), req)
			if tc.ExpectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
