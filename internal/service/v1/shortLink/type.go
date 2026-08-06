package shortLink

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ShortLinkService")

type ShortLinkService struct {
	config *config.Config
	logger logger.ILogger
	repo   repository.IShortLinkRepository
}

func NewShortLinkService(logger logger.ILogger, repo repository.IShortLinkRepository) service.IShortLinkService {
	return &ShortLinkService{
		logger: logger,
		repo:   repo,
	}
}

func WithConfig(svc service.IShortLinkService, config *config.Config) {
	svc.(*ShortLinkService).config = config
}
