package countryRepository

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

type countryRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

var (
	tableName  = "countries"
	otelTracer = otel.Tracer("CountryRepository")
)

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.ICountryRepository {
	return &countryRepository{db, logger}
}
