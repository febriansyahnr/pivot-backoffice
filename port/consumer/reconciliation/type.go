package reconciliation

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ReconciliationConsumer")

type ReconciliationController struct {
	reconSvc service.IReconciliationService
	logger   logger.ILogger
}

func New(logger logger.ILogger, reconSvc service.IReconciliationService) consumer.IReconciliationConsumer {
	return &ReconciliationController{reconSvc, logger}
}
