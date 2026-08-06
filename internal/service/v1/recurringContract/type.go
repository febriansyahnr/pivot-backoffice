package recurringContractService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	port "github.com/paper-indonesia/pivot-backoffice/internal/service"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RecurringContractService")

type service struct {
	log         logger.ILogger
	repo        repository.IRecurringContractRepository
	customerSvc port.ICustomerService
}

func New(
	log logger.ILogger, repo repository.IRecurringContractRepository, customerSvc port.ICustomerService,
) port.IRecurringContractService {
	return &service{log, repo, customerSvc}
}
