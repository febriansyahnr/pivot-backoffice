package country

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V1CountryController")

type V1CountryController struct {
	validate       *validator.Validate
	countryService service.ICountryService
}

func New(
	countryService service.ICountryService,
	validate *validator.Validate,
) controller.V1CountryController {
	return &V1CountryController{
		validate:       validate,
		countryService: countryService,
	}
}
