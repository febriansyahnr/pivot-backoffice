package disbursementConsumerController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("DisbursementConsumer")

type DisbursementConsumer struct {
	logger             logger.ILogger
	disbursementSvc    service.IDisbursementService
	refundProcessorSvc service.IRefundProcessorService
}

func New(
	logger logger.ILogger, disbursementSvc service.IDisbursementService, refundProcessorSvc service.IRefundProcessorService,
) consumer.IDisbursementConsumer {
	return &DisbursementConsumer{logger, disbursementSvc, refundProcessorSvc}
}
