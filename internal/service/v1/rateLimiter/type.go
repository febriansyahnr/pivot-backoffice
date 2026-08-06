package ratelimiter

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const (
	LockDuration                   = 1 * time.Hour
	MaxFailedAttempt               = 3
	FailedAttemptTimeFrameInMinute = 15
	RateLimiterKeyFormat           = "backend-portal:rate-limiter:%s:%s"
	RateLimiterLockKeyFormat       = "backend-portal:rate-limiter:%s:%s:lock"
	MerchantRateLimitKeyConfig     = "backend-portal:rate-limiter:merchants:%s:config"  // $1 is MerchantID
	MerchantRateLimitKey           = "backend-portal:rate-limiter:merchants:%s:%s"      // $1 is MerchantID $1 is configID
	MerchantRateLimitDefaultKey    = "backend-portal:rate-limiter:merchants:%s:default" // $1 is MerchantID

	Pin                   = "pin"
	Password              = "password"
	ExclusiveLockDuration = 5 * time.Second
)

var otelTracer = otel.Tracer("RateLimiterService")

type rateLimitServiceOption func(*rateLimiterService)

type rateLimiterService struct {
	config          *config.Config
	logger          logger.ILogger
	redis           redisExt.IRedisExt
	rateLimiterRepo repository.IRateLimiterRepository
	limiter         redisExt.ILimiter
}

func New(
	logger logger.ILogger,
	redis redisExt.IRedisExt,
	rateLimiterRepo repository.IRateLimiterRepository,
	opts ...rateLimitServiceOption,
) *rateLimiterService {
	service := &rateLimiterService{
		logger:          logger,
		redis:           redis,
		rateLimiterRepo: rateLimiterRepo,
	}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func WithRedisLimiter(limiter redisExt.ILimiter) rateLimitServiceOption {
	return func(s *rateLimiterService) {
		s.limiter = limiter
	}
}

func WithConfig(cfg *config.Config) rateLimitServiceOption {
	return func(s *rateLimiterService) {
		s.config = cfg
	}
}
