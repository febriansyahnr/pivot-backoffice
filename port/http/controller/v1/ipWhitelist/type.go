package ipWhitelistController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("IPWhitelistConfigurationController")

type IPWhitelistConfigurationController struct {
	svc       service.IIPWhitelistService
	validator *validator.Validate
}

func New(svc service.IIPWhitelistService, validator *validator.Validate) controller.V1IPWhitelistController {
	return &IPWhitelistConfigurationController{svc: svc, validator: validator}
}
