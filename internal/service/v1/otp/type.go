package otp

import (
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	jwtExt "github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const otpMaxDigits uint32 = 6
const exclusiveLockDuration = 10 * time.Second

var loc, _ = time.LoadLocation(constant.TimeLoc)
var otelTracer = otel.Tracer("OTPService")

type service struct {
	config            *config.Config
	logger            logger.ILogger
	redis             redisExt.IRedisExt
	jwt               jwtExt.IJwt
	rmq               rabbitMqExt.IRabbitMQExt
	userRepo          repository.IUserRepository
	userEncryptionKey vault.IVaultKeyValue
	generator         services.IOTPGenerator
	limiter           redisExt.ILimiter
	totpRateLimit     config.TOTPRateLimitConfig
}

func New(
	config *config.Config,
	logger logger.ILogger,
	redis redisExt.IRedisExt,
	jwt jwtExt.IJwt,
	rmq rabbitMqExt.IRabbitMQExt,
	userRepo repository.IUserRepository,
	limiter redisExt.ILimiter,
) *service {
	s := &service{
		config:        config,
		logger:        logger,
		redis:         redis,
		jwt:           jwt,
		rmq:           rmq,
		userRepo:      userRepo,
		limiter:       limiter,
		totpRateLimit: config.MultiFactorAuth.TimeBasedOTP.TOTPRateLimit,
	}
	s.generator = s

	if s.totpRateLimit.RequestLimit <= 0 {
		s.totpRateLimit.RequestLimit = 1
	}
	if s.totpRateLimit.RequestWindow <= 0 {
		s.totpRateLimit.RequestWindow = time.Second
	}
	return s
}

func (s *service) WithGenerator(gen services.IOTPGenerator) {
	s.generator = gen
}

func (s *service) WithUserEncryptionKey(encryptionKey vault.IVaultKeyValue) {
	s.userEncryptionKey = encryptionKey
}

func redisOTPKey(identifier string, feature constant.OTPIdentifier, addition ...string) (key string) {
	key = fmt.Sprintf(
		constant.OTPKeyFormatting, identifier, feature.FeatureName(),
	)
	for _, s := range addition {
		key += s
	}
	return
}
