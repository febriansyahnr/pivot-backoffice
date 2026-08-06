package crmProductController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMProductController")

type CRMProductController struct {
	validate       *validator.Validate
	productService service.IProductService
}

func New(productService service.IProductService, validate *validator.Validate) controller.V1CRMProductController {
	return &CRMProductController{
		productService: productService,
		validate:       validate,
	}
}
