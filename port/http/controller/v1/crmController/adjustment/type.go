package adjustment

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

type handler struct {
	validator *validatorExt.Validate
	service   service.IAdjustmentService
}

var otelTracer = otel.Tracer("AdjustmentController")

func New(service service.IAdjustmentService) controller.V1CRMAdjustment {
	return &handler{
		service:   service,
		validator: validatorExt.New(),
	}
}
