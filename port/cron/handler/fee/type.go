package feeHandler

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/port/cron/handler"

	"go.opentelemetry.io/otel"
)

var (
	otelTracer = otel.Tracer("FeeCronHandler")
	tz, _      = time.LoadLocation(constant.TimeLoc)
)

type feeHandler struct {
	logger  logger.ILogger
	service service.IFeeService
}

func New(logger logger.ILogger, service service.IFeeService) handler.IFeeHandler {
	return &feeHandler{logger, service}
}
