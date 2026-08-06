package qris

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"

	"github.com/go-playground/validator/v10"
)

var otelTracer = otel.Tracer("QrisController")

type handler struct {
	validator *validator.Validate
	service   service.IQrisService
	config    *config.Config
}

func New(vld *validatorExt.Validate, service service.IQrisService, cfg *config.Config) controller.V1QrisController {
	return &handler{
		validator: vld,
		service:   service,
		config:    cfg,
	}
}
