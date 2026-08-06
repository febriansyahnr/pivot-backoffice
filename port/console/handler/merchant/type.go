package merchantHandler

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/console"
	"github.com/paper-indonesia/pdk/v2/logger"

	"go.opentelemetry.io/otel"
)

var (
	otelTracer = otel.Tracer("MerchantConsoleHandler")
	tz, _      = time.LoadLocation(constant.TimeLoc)
)

type handler struct {
	logger       logger.ILogger
	service      service.IMerchantService
	reportingSvc service.IReportingService
}

func New(log logger.ILogger, svc service.IMerchantService, reportingSvc service.IReportingService) console.IMerchantCommand {
	return &handler{service: svc, logger: log, reportingSvc: reportingSvc}
}
