package transfer

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer        = otel.Tracer("PaymentController")
	allowedSortOrders = map[string]bool{"ASC": true, "DESC": true}
	allowedSort       = map[string]bool{"createdAt": true, "recipientName": true, "senderName": true, "amount": true}
)

type OptionalTransferFunc func(*TransferController)

type TransferController struct {
	config          *config.Config
	validate        *validator.Validate
	monitor         *monitoring.Monitor
	logger          logger.ILogger
	transferService service.ITransferService
	merchantService service.IMerchantService
}

func New(
	config *config.Config,
	validate *validator.Validate,
	monitor *monitoring.Monitor,
	transferservice service.ITransferService,
	depends ...OptionalTransferFunc,
) controller.V1TransferController {
	c := &TransferController{
		config:          config,
		validate:        validate,
		monitor:         monitor,
		transferService: transferservice,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithMerchantService(service service.IMerchantService) OptionalTransferFunc {
	return func(c *TransferController) {
		c.merchantService = service
	}
}
