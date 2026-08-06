package aml

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMFraudRuleController")

type CRMAmlController struct {
	validate   *validator.Validate
	amlService service.IAmlService
}

func New(
	amlService service.IAmlService,
	validate *validator.Validate,
) controller.V1CRMAmlController {
	return &CRMAmlController{
		validate:   validate,
		amlService: amlService,
	}
}
