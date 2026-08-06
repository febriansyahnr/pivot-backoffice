package recurringContractHandler

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V1InternalRecurringContractController")

type handler struct {
	validate *validatorExt.Validate
	service  service.IRecurringContractService
}

func New(vld *validatorExt.Validate, service service.IRecurringContractService) controller.V1InternalRecurringContractController {
	return &handler{vld, service}
}
