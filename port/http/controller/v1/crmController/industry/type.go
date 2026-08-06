package industry

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMIndustryController")

type CRMIndustryController struct {
	validate        *validator.Validate
	industryService service.IIndustryService
}

func NewController(
	industryService service.IIndustryService,
	validate *validator.Validate,
) controller.V1CRMIndustryController {
	return &CRMIndustryController{
		validate:        validate,
		industryService: industryService,
	}
}
