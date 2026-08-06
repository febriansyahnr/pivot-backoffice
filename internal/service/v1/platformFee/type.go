package platformFeeService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

type PlatformFeeService struct {
	ledgerSvc  service.ILedgerService
	feeSvc     service.IFeeService
	accountSvc service.IAccountService
	logger     logger.ILogger
}

var otelTracer = otel.Tracer("PlatformFeeService")

func New(
	logger logger.ILogger,
	ledgerSvc service.ILedgerService,
	feeSvc service.IFeeService,
	accountSvc service.IAccountService,
) service.IPlatformFeeService {
	return &PlatformFeeService{
		logger:     logger,
		ledgerSvc:  ledgerSvc,
		feeSvc:     feeSvc,
		accountSvc: accountSvc,
	}
}
