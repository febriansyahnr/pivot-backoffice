package industry

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("IndustryRepository")

type repository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) port.IIndustryRepository {
	return &repository{db, logger}
}

const (
	industriesTableName = "industries"
)
