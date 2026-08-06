package merchantTopUp

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MerchantTopUpController")

type handler struct {
	validate *validatorExt.Validate
	service  service.IMerchantTopUpService
}

func New(vld *validatorExt.Validate, svc service.IMerchantTopUpService) controller.V1MerchantTopUpController {
	return &handler{vld, svc}
}
