package liveFeature

import (
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/liveFeature"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("LiveFeatureService")

type LiveFeatureService struct {
	logger         logger.ILogger
	repo           repository.ILiveFeatureRepository
	rabbitMqExt    rabbitMqExt.IRabbitMQExt
	currentVersion liveFeature.AppVersion
	mu             sync.Mutex // Protects access to currentVersion
}

func New(
	logger logger.ILogger,
	repo repository.ILiveFeatureRepository,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	depends ...LiveFeatureServiceFunc,
) service.ILiveFeaturesService {
	return &LiveFeatureService{
		logger:      logger,
		repo:        repo,
		rabbitMqExt: rabbitMqExt,
	}
}

type LiveFeatureServiceFunc func(*LiveFeatureService)
