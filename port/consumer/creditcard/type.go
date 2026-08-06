package creditcard

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CreditCardConsumer")

type CreditCardConsumer struct {
	logger            logger.ILogger
	creditcardSvc     service.ICreditCardService
	orchestratorSvc   service.IOrchestratorService
	paymentSvc        service.IPaymentService
	unifiedPaymentSvc service.IUnifiedPaymentService
	refundSvc         service.IRefundService
}

func New(
	logger logger.ILogger,
	creditcardSvc service.ICreditCardService,
	orchestratorSvc service.IOrchestratorService,
	paymentSvc service.IPaymentService,
	unifiedPaymentSvc service.IUnifiedPaymentService,
	refundSvc service.IRefundService,
) consumer.ICreditCardService {
	return &CreditCardConsumer{
		logger:            logger,
		creditcardSvc:     creditcardSvc,
		orchestratorSvc:   orchestratorSvc,
		paymentSvc:        paymentSvc,
		unifiedPaymentSvc: unifiedPaymentSvc,
		refundSvc:         refundSvc,
	}
}
