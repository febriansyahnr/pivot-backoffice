package refund

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RefundController")

type RefundController struct {
	refundService service.IRefundService
	validate      *validatorExt.Validate
	logger        logger.ILogger
}

type ControllerFunc func(*RefundController)

func New(depends ...ControllerFunc) controller.V1RefundController {
	c := &RefundController{
		validate: validatorExt.New(),
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithLogger(log logger.ILogger) ControllerFunc {
	return func(c *RefundController) {
		c.logger = log
	}
}

func WithRefundService(svc service.IRefundService) ControllerFunc {
	return func(c *RefundController) {
		c.refundService = svc
	}
}
