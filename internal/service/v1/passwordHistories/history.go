package passwordHistories

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PasswordHistoriesService")

type PasswordHistoriesService struct {
	repo   repository.IPasswordHistoriesRepository
	logger logger.ILogger
}

func New(repo repository.IPasswordHistoriesRepository, logger logger.ILogger) service.IPasswordHistoriesService {
	return &PasswordHistoriesService{
		repo:   repo,
		logger: logger,
	}
}
