package purpose

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V1MasterPurposeController")

type Controller struct {
	config *config.Config
}

func New(config *config.Config) controller.V1MasterPurposeController {
	return &Controller{
		config: config,
	}
}
