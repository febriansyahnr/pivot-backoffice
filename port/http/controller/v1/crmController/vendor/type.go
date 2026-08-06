package vendor

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMVendorController")

type CRMVendorController struct {
	vendorService service.IVendorService
	validate      *validator.Validate
}

func New(vendorService service.IVendorService, validate *validator.Validate) controller.V1CRMVendorController {
	return &CRMVendorController{
		vendorService: vendorService,
		validate:      validate,
	}
}
