package customerService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CustomerService")

type CustomerService struct {
	customerRepo   repository.ICustomerRepository
	accountService service.IAccountService
	logger         logger.ILogger
}

func New(customerRepo repository.ICustomerRepository, accountService service.IAccountService, logger logger.ILogger) *CustomerService {
	return &CustomerService{
		customerRepo:   customerRepo,
		accountService: accountService,
		logger:         logger,
	}
}
