package transferRepository

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

type transferRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

var (
	tableName  = "transfers"
	otelTracer = otel.Tracer("TransferRepository")
)

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.ITransferRepository {
	return &transferRepository{
		db:     db,
		logger: logger,
	}
}
