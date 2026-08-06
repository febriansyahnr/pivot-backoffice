package payment

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/port/cron/handler"

	"go.opentelemetry.io/otel"
)

var (
	localTZ, _ = time.LoadLocation(constant.TimeLoc)
	otelTracer = otel.Tracer("PaymentCronHandler")
)

type paymentCronHandler struct {
	logger         logger.ILogger
	paymentService service.IPaymentService
}

func New(logger logger.ILogger, paymentService service.IPaymentService) handler.IPaymentHandler {
	return &paymentCronHandler{
		logger:         logger,
		paymentService: paymentService,
	}

}
