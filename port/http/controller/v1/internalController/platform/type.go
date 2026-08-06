package platformInternalController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V1InternalPlatformController")

type platformController struct {
	config   *config.Config
	validate *validatorExt.Validate
	logger   logger.ILogger

	platformSvc service.IPlatformService
}

func New(
	config *config.Config,
	platformSvc service.IPlatformService,
) controller.V1InternalPlatformController {
	c := &platformController{
		config:      config,
		validate:    validatorExt.New(),
		platformSvc: platformSvc,
	}
	return c
}
