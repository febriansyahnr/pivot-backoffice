package recurringContractRepo

import (
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const tableName = "recurring_contracts"

var otelTracer = otel.Tracer("RecurringContractRepository")

type repository struct {
	log logger.ILogger
	db  mySqlExt.IMySqlExt
}

func New(log logger.ILogger, db mySqlExt.IMySqlExt) port.IRecurringContractRepository {
	return &repository{log, db}
}
