package country

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMCountryController")

type CRMCountryController struct {
	validate       *validator.Validate
	countryService service.ICountryService
}

func New(
	countryService service.ICountryService,
	validate *validator.Validate,
) controller.V1CRMCountryController {
	return &CRMCountryController{
		validate:       validate,
		countryService: countryService,
	}
}
