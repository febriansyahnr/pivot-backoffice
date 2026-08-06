package v1CRMRefundController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V1CRMRefundController")

type refundController struct {
	config   *config.Config
	validate *validatorExt.Validate
	logger   logger.ILogger

	refundSvc services.IRefundService
}

type ControllerFunc func(*refundController)

func New(
	config *config.Config,
	options ...ControllerFunc,
) controller.V1CRMRefundController {
	c := &refundController{
		config:   config,
		validate: validatorExt.New(),
	}
	for _, fn := range options {
		fn(c)
	}
	return c
}

func WithLogger(log logger.ILogger) ControllerFunc {
	return func(c *refundController) {
		c.logger = log
	}
}

func WithRefundService(service services.IRefundService) ControllerFunc {
	return func(c *refundController) {
		c.refundSvc = service
	}
}
