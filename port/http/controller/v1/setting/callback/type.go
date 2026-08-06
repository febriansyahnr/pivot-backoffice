package callback

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CallbackSettingController")

type handler struct {
	validate *validatorExt.Validate
	secret   *config.SecuritySecret
	service  service.ICallbackService
}

func New(validate *validatorExt.Validate, secret *config.SecuritySecret, service service.ICallbackService) controller.V1CallbackSettingController {
	return &handler{validate, secret, service}
}
