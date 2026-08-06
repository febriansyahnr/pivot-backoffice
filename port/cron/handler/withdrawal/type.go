package withdrawalHandler

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	port "github.com/paper-indonesia/pivot-backoffice/port/cron/handler"

	"go.opentelemetry.io/otel"
)

var (
	otelTracer = otel.Tracer("WithdrawalCron")
	tz, _      = time.LoadLocation(constant.TimeLoc)
)

type handler struct {
	logger  logger.ILogger
	service service.IWithdrawalService
}

func New(log logger.ILogger, service service.IWithdrawalService) port.IWithdrawalHandler {
	return &handler{logger: log, service: service}
}
