package reportingConsumer

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ReportingConsumer")

type handler struct {
	logger  logger.ILogger
	service service.IReportingService
}

func New(log logger.ILogger, svc service.IReportingService) consumer.IReportingConsumer {
	return &handler{logger: log, service: svc}
}
