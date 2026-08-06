package crmRateLimiterController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMRateLimiterController")

type CRMRateLimiterController struct {
	svc       service.IRateLimiter
	validator *validator.Validate
}

func New(svc service.IRateLimiter, validator *validator.Validate) controller.V1CRMRateLimiterController {
	return &CRMRateLimiterController{svc: svc, validator: validator}
}
