package platform

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PlatformController")

type PlatformController struct {
	validate    *validator.Validate
	platformSvc service.IPlatformService
	logger      logger.ILogger
}

func New(
	logger logger.ILogger,
	validate *validator.Validate,
	platformSvc service.IPlatformService,
) controller.V1PlatformController {
	return &PlatformController{
		logger:      logger,
		validate:    validate,
		platformSvc: platformSvc,
	}
}
