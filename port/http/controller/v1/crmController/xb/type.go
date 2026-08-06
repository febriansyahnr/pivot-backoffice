package crmXbController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMXbController")

type handler struct {
	validator   *validatorExt.Validate
	XbPayoutSvc service.IXbPayoutService
	logger      logger.ILogger
}

type dependFunc func(c *handler)

func New(xbPayoutSvc service.IXbPayoutService, depends ...dependFunc) controller.V1CRMXbController {
	c := &handler{
		XbPayoutSvc: xbPayoutSvc,
		validator:   validatorExt.New(),
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithLogger(logger logger.ILogger) dependFunc {
	return func(c *handler) {
		c.logger = logger
	}
}
