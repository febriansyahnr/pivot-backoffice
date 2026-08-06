package industry

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("IndustryService")

type IndustryService struct {
	logger logger.ILogger
	repo   repository.IIndustryRepository
}

// NewIndustryService creates a new industry service instance
func NewIndustryService(repo repository.IIndustryRepository, logger logger.ILogger) *IndustryService {
	return &IndustryService{
		repo:   repo,
		logger: logger,
	}
}
