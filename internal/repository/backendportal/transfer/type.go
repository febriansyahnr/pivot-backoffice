package transferRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
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
