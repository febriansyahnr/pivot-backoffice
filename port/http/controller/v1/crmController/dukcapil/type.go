package dukcapil

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMDukcapilController")

type CRMDukcapilController struct {
	dukcapilService service.IDukcapilService
	validate        *validator.Validate
}

type ControllerFunc func(*CRMDukcapilController)

func New(
	dukcapilService service.IDukcapilService,
	validate *validator.Validate,
	depends ...ControllerFunc) *CRMDukcapilController {
	c := &CRMDukcapilController{
		dukcapilService: dukcapilService,
		validate:        validate,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}