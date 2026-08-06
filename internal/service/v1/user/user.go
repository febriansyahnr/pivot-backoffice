package user

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("UserService")

type UserService struct {
	config                 *config.Config
	secret                 *config.Secret
	logger                 logger.ILogger
	userRepo               repository.IUserRepository
	phRepo                 repository.IPasswordHistoriesRepository
	userLoggedInDeviceRepo repository.IUserLoggedInDeviceRepository
	JWT                    jwt.IJwt
	redis                  redisExt.IRedisExt
	otpSvc                 service.IOTP
	rabbitMqExt            rabbitMqExt.IRabbitMQExt
	rateLimiter            service.IRateLimiter
	userLoggedInDeviceSvc  service.IUserLoggedInDeviceService
	permissionSvc          service.IPermissionService
	limiter                redisExt.ILimiter
	roleSvc                service.IRoleService
	userRoleSvc            service.IUserRoleService
	encryptionKey          vault.IVaultKeyValue
	tncSvc                 service.ITNCService
}

type UserServiceFunc func(*UserService)

func WithJWT(jwt jwt.IJwt) UserServiceFunc {
	return func(us *UserService) {
		us.JWT = jwt
	}
}

func WithRedisClient(rdb redisExt.IRedisExt) UserServiceFunc {
	return func(us *UserService) {
		us.redis = rdb
	}
}

func WithRateLimiter(rl service.IRateLimiter) UserServiceFunc {
	return func(us *UserService) {
		us.rateLimiter = rl
	}
}

func WithOTPService(otp service.IOTP) UserServiceFunc {
	return func(us *UserService) {
		us.otpSvc = otp
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) UserServiceFunc {
	return func(us *UserService) {
		us.rabbitMqExt = rmq
	}
}

func WithUserLoggedInDeviceService(svc service.IUserLoggedInDeviceService) UserServiceFunc {
	return func(us *UserService) {
		us.userLoggedInDeviceSvc = svc
	}
}

func WithPermissionService(svc service.IPermissionService) UserServiceFunc {
	return func(us *UserService) {
		us.permissionSvc = svc
	}
}

func WithLimiter(limiter redisExt.ILimiter) UserServiceFunc {
	return func(s *UserService) {
		s.limiter = limiter
	}
}

func WithRoleService(svc service.IRoleService) UserServiceFunc {
	return func(us *UserService) {
		us.roleSvc = svc
	}
}

func WithUserRoleService(svc service.IUserRoleService) UserServiceFunc {
	return func(us *UserService) {
		us.userRoleSvc = svc
	}
}

func WithUserLoggedInDeviceRepo(repo repository.IUserLoggedInDeviceRepository) UserServiceFunc {
	return func(s *UserService) {
		s.userLoggedInDeviceRepo = repo
	}
}

func WithEncryptionKey(kv vault.IVaultKeyValue) UserServiceFunc {
	return func(us *UserService) {
		us.encryptionKey = kv
	}
}

func WithTNCService(s service.ITNCService) UserServiceFunc {
	return func(us *UserService) {
		us.tncSvc = s
	}
}

func New(
	config *config.Config,
	secret *config.Secret,
	logger logger.ILogger,
	repo repository.IUserRepository,
	ph repository.IPasswordHistoriesRepository,
	depends ...UserServiceFunc,
) service.IUserService {
	s := &UserService{
		config:   config,
		secret:   secret,
		userRepo: repo,
		logger:   logger,
		phRepo:   ph,
	}
	for _, fn := range depends {
		fn(s)
	}
	return s
}

func redisTokenKey(identifier string, feature constant.UserFeatureIdentifier, addition ...string) (key string) {
	key = fmt.Sprintf(
		constant.UserKeyFormatting, identifier, feature.FeatureName(),
	)

	if identifier == "" {
		key = fmt.Sprintf(
			constant.UserKeyWithoutIdentifierFormatting, feature.FeatureName(),
		)
	}

	for _, s := range addition {
		key += s
	}
	return
}

func (s *UserService) ActivityLog(ctx context.Context, merchantID, userID *string, tag string, activity string, params map[string]string) {
	_ = s.rabbitMqExt.PublishActivity(ctx, merchantID, userID, tag, activity, params)
}
