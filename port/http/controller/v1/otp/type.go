package otp

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"

	validator "github.com/go-playground/validator/v10"
)

var otelTracer = otel.Tracer("OtpController")

type handler struct {
	validate *validator.Validate
	service  service.IOTP
	userSvc  service.IUserService
}

func New(service service.IOTP, userSvc service.IUserService) *handler {
	return &handler{validator.New(), service, userSvc}
}
