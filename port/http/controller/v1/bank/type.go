package bankController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("BankController")

type Controller struct {
	config *config.Config
}

func New(config *config.Config) controller.V1BankController {
	return &Controller{
		config: config,
	}
}
