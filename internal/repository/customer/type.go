package customerRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const tableName = "customers"

var otelTracer = otel.Tracer("CustomerRepository")

type CustomerRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(
	db mySqlExt.IMySqlExt,
	logger logger.ILogger,
) repository.ICustomerRepository {
	return &CustomerRepository{
		db:     db,
		logger: logger,
	}
}
