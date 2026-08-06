package location

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("LocationController")

type handler struct {
	validate *validatorExt.Validate
	service  service.IAddrLocationService
}

func New(vld *validatorExt.Validate, service service.IAddrLocationService) controller.V1AddrLocationController {
	return &handler{
		validate: vld,
		service:  service,
	}
}
