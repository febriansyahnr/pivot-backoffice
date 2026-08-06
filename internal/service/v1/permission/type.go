package permissionService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PermissionService")

type PermissionService struct {
	repo   repository.IPermissionRepository
	logger logger.ILogger
	redis  redisExt.IRedisExt
}

type PermissionServiceFunc func(*PermissionService)

func New(repo repository.IPermissionRepository, logger logger.ILogger, depends ...PermissionServiceFunc) service.IPermissionService {
	s := &PermissionService{
		repo:   repo,
		logger: logger,
	}

	for _, fn := range depends {
		fn(s)
	}
	return s
}

func WithRedisClient(rdb redisExt.IRedisExt) PermissionServiceFunc {
	return func(us *PermissionService) {
		us.redis = rdb
	}
}
