package tnc

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("TNCSigningController")

type TNCSigningController struct {
	service  service.ITNCService
	validate *validator.Validate
}

func New(
	service service.ITNCService,
	validate *validator.Validate,
) controller.V1TNCSigningController {
	return &TNCSigningController{
		service:  service,
		validate: validate,
	}
}
