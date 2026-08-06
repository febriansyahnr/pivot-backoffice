package customer

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMCustomerController")

type CRMCustomerController struct {
	validate        *validator.Validate
	customerService service.ICustomerService
}

func New(
	customerService service.ICustomerService,
	validate *validator.Validate,
) controller.V1CRMCustomerController {
	return &CRMCustomerController{
		validate:        validate,
		customerService: customerService,
	}
}