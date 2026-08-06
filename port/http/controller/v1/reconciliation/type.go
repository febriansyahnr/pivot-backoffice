package reconciliation

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ReconciliationController")

type ReconciliationController struct {
	validate *validator.Validate
	reconSvc service.IReconciliationService
	logger   logger.ILogger
}

func New(
	logger logger.ILogger,
	validate *validator.Validate,
	reconSvc service.IReconciliationService,
) controller.V1ReconciliationController {
	return &ReconciliationController{
		validate: validate,
		logger:   logger,
		reconSvc: reconSvc,
	}
}
