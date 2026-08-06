package activityController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ActivityController")

type activity struct {
	config      *config.Config
	activitySvc service.IActivityService
	validate    *validatorExt.Validate
}

func New(
	config *config.Config,
	validate *validatorExt.Validate,
	activitySvc service.IActivityService,
) controller.V1ActivityController {
	return &activity{
		validate: validate, config: config, activitySvc: activitySvc,
	}
}
