package callbackController

import (
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
)

var otelTracer = otel.Tracer("CallbackController")

type CallbackController struct {
	config      *config.Config
	validate    *validator.Validate
	callbackSvc service.ICallbackService
	rabbitMqExt rabbitMqExt.IRabbitMQExt
}

func New(
	config *config.Config,
	validate *validator.Validate,
	callbackSvc service.ICallbackService,
	rabbitMqExt rabbitMqExt.IRabbitMQExt) controller.V1CallbackController {
	return &CallbackController{
		config:      config,
		validate:    validate,
		callbackSvc: callbackSvc,
		rabbitMqExt: rabbitMqExt,
	}
}
