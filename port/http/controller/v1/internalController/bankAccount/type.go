package internalBankAccountController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalBankAccountController")

type InternalBankAccountController struct {
	svc service.IBankAccountService
}

func New(
	svc service.IBankAccountService,
) controller.V1InternalBankAccountController {
	return &InternalBankAccountController{
		svc: svc,
	}
}
