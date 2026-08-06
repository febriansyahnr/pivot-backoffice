package customerController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalCustomerController")

type V1InternalCustomerController struct {
	customerService service.ICustomerService
	validate        *validator.Validate
}

func New(customerService service.ICustomerService,
	validate *validator.Validate,
) *V1InternalCustomerController {
	return &V1InternalCustomerController{
		customerService: customerService,
		validate:        validate,
	}
}
