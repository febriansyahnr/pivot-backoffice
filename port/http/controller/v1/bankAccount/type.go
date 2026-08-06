package bankAccount

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

type handler struct {
	validator *validatorExt.Validate
	service   service.IBankAccountService
}

var otelTracer = otel.Tracer("BankAccountController")

func New(vld *validatorExt.Validate, svc service.IBankAccountService) controller.V1BankAccountController {
	return &handler{
		validator: vld, service: svc,
	}
}
