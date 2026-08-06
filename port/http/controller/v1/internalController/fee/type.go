package internalFeeController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalFeeController")

type InternalFeeController struct {
	svc service.IFeeService
}

func New(
	svc service.IFeeService,
) controller.V1InternalFeeController {
	return &InternalFeeController{
		svc: svc,
	}
}
