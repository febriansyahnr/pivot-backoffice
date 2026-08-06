package ipwhitelistService

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("IPWhitelistService")

type IPWhitelistService struct {
	config        *config.Config
	logger        logger.ILogger
	whitelistRepo repository.IIPWhitelistRepository
	redisCache    redisExt.IRedisExt
}

type IPWhitelistServiceOpt func(*IPWhitelistService)

func New(
	logger logger.ILogger,
	whitelistRepo repository.IIPWhitelistRepository,
	redisCache redisExt.IRedisExt,
	opts ...IPWhitelistServiceOpt,
) service.IIPWhitelistService {
	s := &IPWhitelistService{
		logger:        logger,
		whitelistRepo: whitelistRepo,
		redisCache:    redisCache,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithConfig(cfg *config.Config) IPWhitelistServiceOpt {
	return func(s *IPWhitelistService) {
		s.config = cfg
	}
}
