package country

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

type countryService struct {
	repo   repository.ICountryRepository
	logger logger.ILogger
}

var otelTracer = otel.Tracer("CountryService")

func New(
	repo repository.ICountryRepository,
	logger logger.ILogger,
) service.ICountryService {
	svc := &countryService{
		repo:   repo,
		logger: logger,
	}
	return svc
}
